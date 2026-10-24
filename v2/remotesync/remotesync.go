// Package remotesync keeps this machine's users in step with the shared remote
// db (the MastersClosetRemoteDB server at remote_host_url). It speaks v1's wire
// protocol unchanged — GET /changed, GET /download, POST /import with
// <prefix>-API-KEY / -UUID / -CLIENT-ID / -SEQUENCE-ID headers — but fixes v1's
// loop: changes are applied in sequence order, the pull cursor only advances
// past changes actually applied (v1 jumped it to its own upload's sequence and
// silently skipped other clients' edits), concurrent edits to the same user are
// merged instead of overwritten, and no HTTP call runs inside a bolt
// transaction.
package remotesync

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/0187773933/MastersCloset/v2/config"
	"github.com/0187773933/MastersCloset/v2/logger"
	"github.com/0187773933/MastersCloset/v2/user"
	bolt "github.com/boltdb/bolt"
)

// Interval between sync rounds (v1 used the same).
var Interval = 12 * time.Second

const (
	// v1's bucket/key for the pull cursor, kept so a machine that last ran v1
	// resumes from where it left off.
	bucketMisc  = "MISC"
	keyCursor   = "remote-last-sequence"
	bucketOwn   = "remote-own-sequences" // sequences of our own uploads, skipped on pull
	maxResponse = 64 << 20
)

// Status is a snapshot of the most recent sync round, for the admin dashboard.
type Status struct {
	Enabled     bool      `json:"enabled"`
	Running     bool      `json:"running"`
	LastAttempt time.Time `json:"last_attempt"`
	LastSuccess time.Time `json:"last_success"`
	LastError   string    `json:"last_error"`
	Pending     int       `json:"pending"`
	Pulled      int       `json:"pulled"` // in the last round
	Merged      int       `json:"merged"` // in the last round
	Pushed      int       `json:"pushed"` // in the last round
	TotalPulled int       `json:"total_pulled"`
	TotalPushed int       `json:"total_pushed"`
}

// Syncer runs the pull/push loop.
type Syncer struct {
	db          *bolt.DB
	cfg         *config.Manager
	store       *user.Store
	fingerprint string
	http        *http.Client
	ctx         context.Context // cancelled on shutdown; aborts in-flight requests

	round sync.Mutex // one round at a time

	mu     sync.Mutex
	status Status
}

// New wires a syncer. fingerprint identifies this machine to the remote.
func New(db *bolt.DB, cfg *config.Manager, store *user.Store, fingerprint string) *Syncer {
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(bucketMisc))
		tx.CreateBucketIfNotExists([]byte(bucketOwn))
		return nil
	})
	return &Syncer{
		db:          db,
		cfg:         cfg,
		store:       store,
		fingerprint: fingerprint,
		http:        &http.Client{Timeout: 30 * time.Second},
		ctx:         context.Background(),
	}
}

// Start runs a round immediately and then every Interval until ctx ends.
// Start must be called before any SyncNow.
func (s *Syncer) Start(ctx context.Context) {
	s.ctx = ctx
	go func() {
		s.SyncNow()
		t := time.NewTicker(Interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.SyncNow()
			}
		}
	}()
}

// Wait blocks until any round in progress has finished. After cancelling the
// Start context this returns promptly, since requests are aborted.
func (s *Syncer) Wait() {
	s.round.Lock()
	s.round.Unlock()
}

// Status returns the latest round's outcome.
func (s *Syncer) Status() Status {
	s.mu.Lock()
	st := s.status
	s.mu.Unlock()
	st.Enabled = s.enabled(s.cfg.Snapshot())
	st.Pending = s.store.PendingCount()
	return st
}

func (s *Syncer) enabled(c config.Config) bool {
	return strings.TrimSpace(c.RemoteHostURL) != "" && strings.TrimSpace(c.RemoteHostAPIKey) != ""
}

// SyncNow runs one pull-then-push round (waiting for any round in progress).
func (s *Syncer) SyncNow() Status {
	s.round.Lock()
	defer s.round.Unlock()

	c := s.cfg.Snapshot()
	if !s.enabled(c) || s.ctx.Err() != nil {
		return s.Status()
	}
	log := logger.GetLogger()
	s.setStatus(func(st *Status) { st.Running = true; st.LastAttempt = time.Now() })

	pulled, merged, err := s.pull(c)
	pushed := 0
	if err == nil {
		// Only push after a clean pull, so we never upload a user without
		// first merging what other machines sent for them.
		pushed, err = s.push(c)
	}

	s.setStatus(func(st *Status) {
		st.Running = false
		st.Pulled, st.Merged, st.Pushed = pulled, merged, pushed
		st.TotalPulled += pulled
		st.TotalPushed += pushed
		if err != nil {
			st.LastError = err.Error()
		} else {
			st.LastError = ""
			st.LastSuccess = time.Now()
		}
	})
	if pulled > 0 || pushed > 0 {
		log.Info(fmt.Sprintf("remote sync: pulled %d (merged %d), pushed %d, %d still pending", pulled, merged, pushed, s.store.PendingCount()))
	}
	if err != nil {
		log.Info(fmt.Sprintf("remote sync: %v", err))
	}
	return s.Status()
}

func (s *Syncer) setStatus(fn func(*Status)) {
	s.mu.Lock()
	fn(&s.status)
	s.mu.Unlock()
}

type change struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

// pull applies every remote change after our cursor, oldest first.
func (s *Syncer) pull(c config.Config) (pulled, merged int, err error) {
	log := logger.GetLogger()
	cursor := s.getCursor()

	var list struct {
		Changes []change `json:"changes"`
	}
	if err := s.do(c, "GET", "/changed", map[string]string{
		"CLIENT-ID":   s.fingerprint,
		"SEQUENCE-ID": cursor,
	}, nil, &list); err != nil {
		return 0, 0, fmt.Errorf("pull: list changes: %w", err)
	}
	// The server builds the list from a map, so it arrives unordered. IDs are
	// ULIDs: lexical order is time order.
	sort.Slice(list.Changes, func(i, j int) bool { return list.Changes[i].ID < list.Changes[j].ID })

	// /download always returns the current record, so a user changed many
	// times since our cursor (a long-offline machine catching up) only needs
	// fetching once, at its last change.
	last := map[string]int{}
	for i, ch := range list.Changes {
		last[ch.UUID] = i
	}

	for i, ch := range list.Changes {
		if last[ch.UUID] != i {
			continue
		}
		if s.isOwn(ch.ID) {
			s.setCursor(ch.ID)
			continue
		}
		var dl struct {
			UUID      string `json:"uuid"`
			UserBytes []byte `json:"user_bytes"`
		}
		if err := s.do(c, "GET", "/download", map[string]string{"UUID": ch.UUID}, nil, &dl); err != nil {
			return pulled, merged, fmt.Errorf("pull: download %s: %w", ch.UUID, err)
		}
		if len(dl.UserBytes) == 0 {
			// Remote has no body for it (nothing to apply); move on.
			s.setCursor(ch.ID)
			continue
		}
		u, wasMerged, err := s.store.ApplyRemote(dl.UserBytes)
		if err != nil {
			if errors.Is(err, user.ErrUndecryptable) {
				// A record we can never read would block the cursor forever.
				log.Error(fmt.Sprintf("remote sync: skipping change %s for %s: %v", ch.ID, ch.UUID, err))
				s.setCursor(ch.ID)
				continue
			}
			return pulled, merged, fmt.Errorf("pull: apply %s: %w", ch.UUID, err)
		}
		pulled++
		if wasMerged {
			merged++
			log.Info(fmt.Sprintf("remote sync: merged remote and local edits for %s (%s)", u.UUID, u.NameString))
		}
		s.setCursor(ch.ID)
	}
	return pulled, merged, nil
}

// push uploads every pending user.
func (s *Syncer) push(c config.Config) (pushed int, err error) {
	for userUUID, enc := range s.store.PendingUploads() {
		var res struct {
			Result   bool   `json:"result"`
			Sequence string `json:"sequence"`
		}
		if err := s.do(c, "POST", "/import", map[string]string{
			"CLIENT-ID": c.RemoteHostClientID,
			"UUID":      userUUID,
		}, enc, &res); err != nil {
			return pushed, fmt.Errorf("push %s: %w", userUUID, err)
		}
		if !res.Result {
			return pushed, fmt.Errorf("push %s: remote rejected the upload", userUUID)
		}
		s.markOwn(res.Sequence)
		s.store.MarkUploaded(userUUID, enc)
		pushed++
	}
	return pushed, nil
}

// do performs one request against the remote and decodes its JSON reply.
func (s *Syncer) do(c config.Config, method, path string, headers map[string]string, body []byte, out interface{}) error {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(s.ctx, method, strings.TrimRight(c.RemoteHostURL, "/")+path, rdr)
	if err != nil {
		return err
	}
	prefix := c.RemoteHostHeaderPrefix
	if prefix == "" {
		prefix = "MC"
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(prefix+"-API-KEY", c.RemoteHostAPIKey)
	for k, v := range headers {
		req.Header.Set(prefix+"-"+k, v)
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := string(data)
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		return fmt.Errorf("http %d: %s", resp.StatusCode, snippet)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (s *Syncer) getCursor() string {
	var v string
	s.db.View(func(tx *bolt.Tx) error {
		v = string(tx.Bucket([]byte(bucketMisc)).Get([]byte(keyCursor)))
		return nil
	})
	return v
}

func (s *Syncer) setCursor(id string) {
	s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketMisc)).Put([]byte(keyCursor), []byte(id))
	})
}

func (s *Syncer) isOwn(seq string) bool {
	own := false
	s.db.View(func(tx *bolt.Tx) error {
		own = tx.Bucket([]byte(bucketOwn)).Get([]byte(seq)) != nil
		return nil
	})
	return own
}

// markOwn remembers one of our upload sequences. Entries at or behind the pull
// cursor are no longer needed and are pruned.
func (s *Syncer) markOwn(seq string) {
	if seq == "" {
		return
	}
	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketOwn))
		cursor := string(tx.Bucket([]byte(bucketMisc)).Get([]byte(keyCursor)))
		var stale [][]byte
		b.ForEach(func(k, _ []byte) error {
			if cursor != "" && string(k) <= cursor {
				stale = append(stale, append([]byte(nil), k...))
			}
			return nil
		})
		for _, k := range stale {
			b.Delete(k)
		}
		return b.Put([]byte(seq), []byte{1})
	})
}
