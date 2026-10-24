package remotesync

import (
	"context"
	json "encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/0187773933/MastersCloset/v2/config"
	"github.com/0187773933/MastersCloset/v2/logger"
	"github.com/0187773933/MastersCloset/v2/user"
	bolt "github.com/boltdb/bolt"
	ulid "github.com/oklog/ulid/v2"
)

const (
	testKey    = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	testAPIKey = "test-api-key"
)

// fakeRemote mimics MastersClosetRemoteDB: latest bytes per user, plus an
// ordered change log keyed by ULID, returned unordered after the cursor.
type fakeRemote struct {
	mu      sync.Mutex
	users   map[string][]byte
	changes map[string]string // ulid -> uuid
}

func (f *fakeRemote) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Header.Get("MC-API-KEY") != testAPIKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch r.URL.Path {
	case "/import":
		body, _ := io.ReadAll(r.Body)
		seq := ulid.Make().String()
		f.users[r.Header.Get("MC-UUID")] = body
		f.changes[seq] = r.Header.Get("MC-UUID")
		json.NewEncoder(w).Encode(map[string]interface{}{"result": true, "sequence": seq})
	case "/changed":
		since := r.Header.Get("MC-SEQUENCE-ID")
		var keys []string
		for k := range f.changes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := []map[string]string{}
		for _, k := range keys {
			if since == "" || since == "0" || k > since {
				out = append(out, map[string]string{"id": k, "uuid": f.changes[k]})
			}
		}
		// The real server builds this from a map: order is not guaranteed.
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"changes": out})
	case "/download":
		id := r.Header.Get("MC-UUID")
		json.NewEncoder(w).Encode(map[string]interface{}{"uuid": id, "user_bytes": f.users[id]})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

type machine struct {
	store *user.Store
	sync  *Syncer
}

func newMachine(t *testing.T, name, url string) machine {
	t.Helper()
	db, err := bolt.Open(filepath.Join(t.TempDir(), name+".db"), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := config.NewManager(db)
	cfg.LoadOrSeed(config.Config{
		BoltDBEncryptionKey:    testKey,
		TimeZone:               "America/New_York",
		RemoteHostURL:          url,
		RemoteHostAPIKey:       testAPIKey,
		RemoteHostHeaderPrefix: "MC",
		RemoteHostClientID:     name,
	})
	store := user.NewStore(db, cfg, nil)
	s := New(db, cfg, store, name)
	s.ctx = context.Background()
	return machine{store, s}
}

func (m machine) round(t *testing.T) Status {
	t.Helper()
	st := m.sync.SyncNow()
	if st.LastError != "" {
		t.Fatalf("sync error: %s", st.LastError)
	}
	return st
}

func remoteURL(t *testing.T) string {
	if u := os.Getenv("MCT_SYNC_E2E_URL"); u != "" {
		return u // run against a real (scratch!) MastersClosetRemoteDB
	}
	srv := httptest.NewServer(&fakeRemote{users: map[string][]byte{}, changes: map[string]string{}})
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestSyncCatchUpAndMerge(t *testing.T) {
	// Keep the daily log file out of the real ~/.config/mct.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	logger.Init("", "America/New_York")
	url := remoteURL(t)
	a := newMachine(t, "A", url)
	b := newMachine(t, "B", url)

	// A has been running offline: a user and a check-in are waiting in its pool.
	jane, _ := a.store.Create("Jane Doe")
	if _, ok := a.store.CheckInShopping(jane.UUID, user.ShoppingTicket{ShoppingFor: 1, Tops: 1}); !ok {
		t.Fatal("check-in failed")
	}
	if st := a.round(t); st.Pushed != 1 || st.Pending != 0 {
		t.Fatalf("A push: %+v", st)
	}

	// B pulls Jane, check-in included and resolvable by ULID.
	if st := b.round(t); st.Pulled != 1 {
		t.Fatalf("B pull: %+v", st)
	}
	got, ok := b.store.Get(jane.UUID)
	if !ok || len(got.CheckIns) != 1 {
		t.Fatalf("B copy of Jane: ok=%v %+v", ok, got.CheckIns)
	}
	if _, ok := b.store.GetCheckIn(got.CheckIns[0].ULID); !ok {
		t.Fatal("pulled check-in should resolve by ULID on B")
	}
	if st := b.round(t); st.Pulled != 0 || st.Pushed != 0 {
		t.Fatalf("B second round should be idle: %+v", st)
	}
	if st := a.round(t); st.Pulled != 0 {
		t.Fatalf("A must not re-pull its own upload: %+v", st)
	}

	// Both machines check Jane in while out of touch with each other.
	ca, _ := a.store.CheckInShopping(jane.UUID, user.ShoppingTicket{ShoppingFor: 1, Shoes: 1})
	cb, _ := b.store.CheckInShopping(jane.UUID, user.ShoppingTicket{ShoppingFor: 1, Dresses: 1})

	a.round(t)                                              // A pushes its visit
	if st := b.round(t); st.Merged != 1 || st.Pushed != 1 { // B merges A's, pushes merged
		t.Fatalf("B merge round: %+v", st)
	}
	a.round(t) // A pulls the merged record

	for name, m := range map[string]machine{"A": a, "B": b} {
		u, _ := m.store.Get(jane.UUID)
		ids := map[string]bool{}
		for _, c := range u.CheckIns {
			ids[c.ULID] = true
		}
		if len(u.CheckIns) != 3 || !ids[ca.ULID] || !ids[cb.ULID] {
			t.Errorf("%s should have all 3 check-ins, has %d", name, len(u.CheckIns))
		}
		if m.store.PendingCount() != 0 {
			t.Errorf("%s still has %d pending", name, m.store.PendingCount())
		}
	}
}
