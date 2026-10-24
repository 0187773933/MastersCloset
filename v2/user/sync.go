package user

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/0187773933/MastersCloset/v2/encryption"
	bolt "github.com/boltdb/bolt"
)

// This file is the store's half of remote sync (v2/remotesync drives it).
// Every Save with Remote set leaves the latest encrypted record for that user in
// the remote-upload pool; the sync loop drains the pool to the remote db and
// applies pulled records here, merging when a local edit is still pending.

// PendingUploads snapshots the remote-upload pool: user UUID -> encrypted record.
func (s *Store) PendingUploads() map[string][]byte {
	out := map[string][]byte{}
	s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketUpload)).ForEach(func(k, v []byte) error {
			out[string(k)] = append([]byte(nil), v...)
			return nil
		})
	})
	return out
}

// PendingCount is the number of users waiting to be pushed.
func (s *Store) PendingCount() int {
	n := 0
	s.db.View(func(tx *bolt.Tx) error {
		n = tx.Bucket([]byte(bucketUpload)).Stats().KeyN
		return nil
	})
	return n
}

// MarkUploaded drops a pool entry once the remote has accepted it, but only if
// the entry is still the exact bytes that were sent. A local edit that landed
// during the upload stays queued for the next round.
func (s *Store) MarkUploaded(userUUID string, sent []byte) {
	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketUpload))
		if cur := b.Get([]byte(userUUID)); cur != nil && bytes.Equal(cur, sent) {
			return b.Delete([]byte(userUUID))
		}
		return nil
	})
}

// ErrUndecryptable means a pulled record could not be decrypted with this
// machine's bolt_db_encryption_key.
var ErrUndecryptable = errors.New("could not decrypt remote record (encryption key mismatch?)")

// ApplyRemote stores a user record pulled from the remote db. With no local
// edit pending for that user the remote version is taken as-is. With one
// pending, the two are merged (see mergeUsers) and the result is re-queued so
// the remote gets the merged record too. merged reports which path was taken.
func (s *Store) ApplyRemote(enc []byte) (u User, merged bool, err error) {
	remote, ok := s.decode(enc)
	if !ok || remote.UUID == "" {
		return u, false, ErrUndecryptable
	}
	key := []byte(remote.UUID)

	// The guard re-checks, inside the write, that neither the stored record nor
	// the pending entry changed since we read them; otherwise read and retry.
	for attempt := 0; attempt < 5; attempt++ {
		var localRaw, pendingRaw []byte
		s.db.View(func(tx *bolt.Tx) error {
			localRaw = append([]byte(nil), tx.Bucket([]byte(bucketUsers)).Get(key)...)
			pendingRaw = append([]byte(nil), tx.Bucket([]byte(bucketUpload)).Get(key)...)
			return nil
		})
		guard := func(tx *bolt.Tx) bool {
			return bytes.Equal(tx.Bucket([]byte(bucketUsers)).Get(key), localRaw) &&
				bytes.Equal(tx.Bucket([]byte(bucketUpload)).Get(key), pendingRaw)
		}

		u = remote
		opts := SaveOptions{KeepModified: true}
		if len(pendingRaw) > 0 {
			if local, ok := s.decode(localRaw); ok {
				u = mergeUsers(local, remote)
				merged = true
				opts = SaveOptions{Remote: true}
			}
		}
		saved, err := s.saveIf(&u, opts, guard)
		if err != nil {
			return u, merged, err
		}
		if saved {
			return u, merged, nil
		}
		merged = false
	}
	return u, false, fmt.Errorf("user %s kept changing locally; will retry next sync", remote.UUID)
}

// mergeUsers reconciles a local record that has unpushed edits with the
// remote's copy of the same user. Profile fields (identity, family, balance,
// ...) come from whichever side was modified more recently; check-ins and
// barcodes are unioned so a visit recorded on either machine is never lost.
func mergeUsers(local, remote User) User {
	newer, older := local, remote
	if modifiedAt(remote).After(modifiedAt(local)) {
		newer, older = remote, local
	}
	out := newer
	out.CheckIns = mergeCheckIns(newer.CheckIns, older.CheckIns)
	out.FailedCheckIns = mergeCheckIns(newer.FailedCheckIns, older.FailedCheckIns)
	out.Barcodes = mergeStrings(newer.Barcodes, older.Barcodes)
	if older.TotalGuestsAdmitted > out.TotalGuestsAdmitted {
		out.TotalGuestsAdmitted = older.TotalGuestsAdmitted
	}
	return out
}

func modifiedAt(u User) time.Time {
	t, _ := time.Parse("02Jan2006 15:04:05.000", u.ModifiedDate+" "+u.ModifiedTime)
	return t
}

// checkInKey identifies a check-in across machines: its ULID, or for v1-era
// records without one, when it happened and what kind it was.
func checkInKey(ci CheckIn) string {
	if ci.ULID != "" {
		return ci.ULID
	}
	return ci.Date + "|" + ci.Time + "|" + ci.Type
}

func checkInAt(ci CheckIn) time.Time {
	t, _ := time.Parse("02Jan2006 15:04:05.000", ci.Date+" "+ci.Time)
	return t
}

// mergeCheckIns unions two check-in lists (a wins on duplicates) and orders
// the result oldest first; Evaluate treats the last element as the most recent.
func mergeCheckIns(a, b []CheckIn) []CheckIn {
	seen := map[string]bool{}
	var out []CheckIn
	for _, list := range [][]CheckIn{a, b} {
		for _, ci := range list {
			k := checkInKey(ci)
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, ci)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti, tj := checkInAt(out[i]), checkInAt(out[j])
		if !ti.Equal(tj) {
			return ti.Before(tj)
		}
		return out[i].ULID < out[j].ULID
	})
	return out
}

func mergeStrings(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, list := range [][]string{a, b} {
		for _, v := range list {
			if v == "" || seen[v] {
				continue
			}
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// encode encrypts a user record the way Save does (tests use it to fabricate
// records "pulled" from the remote).
func (s *Store) encode(u User) []byte {
	payload, _ := json.Marshal(u)
	return encryption.ChaChaEncryptBytes(s.encryptionKey(), payload)
}
