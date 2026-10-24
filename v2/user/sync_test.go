package user

import (
	"path/filepath"
	"testing"

	"github.com/0187773933/MastersCloset/v2/config"
	bolt "github.com/boltdb/bolt"
)

const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := bolt.Open(filepath.Join(t.TempDir(), "test.db"), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := config.NewManager(db)
	if _, err := cfg.LoadOrSeed(config.Config{BoltDBEncryptionKey: testKey, TimeZone: "America/New_York"}); err != nil {
		t.Fatal(err)
	}
	return NewStore(db, cfg, nil)
}

func ci(ulid, date, tm string) CheckIn {
	return CheckIn{ULID: ulid, Date: date, Time: tm, Type: "normal", Result: true}
}

func TestMergeUnionsCheckInsInOrder(t *testing.T) {
	local := User{UUID: "u", ModifiedDate: "10OCT2026", ModifiedTime: "10:00:00.000",
		CheckIns: []CheckIn{ci("A", "01SEP2026", "10:00:00.000"), ci("C", "20OCT2026", "11:00:00.000")}}
	remote := User{UUID: "u", ModifiedDate: "05OCT2026", ModifiedTime: "10:00:00.000",
		CheckIns: []CheckIn{ci("A", "01SEP2026", "10:00:00.000"), ci("B", "01OCT2026", "09:00:00.000")}}
	m := mergeUsers(local, remote)
	got := ""
	for _, c := range m.CheckIns {
		got += c.ULID
	}
	if got != "ABC" {
		t.Errorf("check-ins = %q, want ABC (union, deduped, oldest first)", got)
	}
}

func TestMergeDedupesLegacyCheckInsWithoutULID(t *testing.T) {
	old := CheckIn{Date: "01JAN2024", Time: "09:00:00.000", Type: "new", Result: true}
	m := mergeUsers(User{CheckIns: []CheckIn{old}}, User{CheckIns: []CheckIn{old}})
	if len(m.CheckIns) != 1 {
		t.Errorf("got %d check-ins, want 1", len(m.CheckIns))
	}
}

func TestMergeNewerProfileWins(t *testing.T) {
	local := User{UUID: "u", PhoneNumber: "local", ModifiedDate: "01OCT2026", ModifiedTime: "10:00:00.000", Barcodes: []string{"1"}}
	remote := User{UUID: "u", PhoneNumber: "remote", ModifiedDate: "02OCT2026", ModifiedTime: "10:00:00.000", Barcodes: []string{"2"}}
	m := mergeUsers(local, remote)
	if m.PhoneNumber != "remote" {
		t.Errorf("phone = %q, want the more recently modified side", m.PhoneNumber)
	}
	if len(m.Barcodes) != 2 {
		t.Errorf("barcodes = %v, want both", m.Barcodes)
	}
}

func TestApplyRemoteWithoutPendingTakesRemote(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.Create("Jane Doe")
	s.MarkUploaded(u.UUID, s.PendingUploads()[u.UUID])

	remote := u
	remote.PhoneNumber = "555"
	remote.CheckIns = []CheckIn{ci("R1", "01OCT2026", "10:00:00.000")}
	if _, merged, err := s.ApplyRemote(s.encode(remote)); err != nil || merged {
		t.Fatalf("merged=%v err=%v, want a plain apply", merged, err)
	}
	got, _ := s.Get(u.UUID)
	if got.PhoneNumber != "555" || len(got.CheckIns) != 1 {
		t.Errorf("remote version not applied: %+v", got)
	}
	if s.PendingCount() != 0 {
		t.Errorf("applying a remote record must not queue it for upload")
	}
	if _, ok := s.GetCheckIn("R1"); !ok {
		t.Errorf("pulled check-in should resolve by ULID")
	}
}

func TestApplyRemoteWithPendingMergesAndRequeues(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.Create("Jane Doe")
	local := u
	local.CheckIns = []CheckIn{ci("L1", "02OCT2026", "10:00:00.000")}
	s.Save(&local, SaveOptions{Remote: true})

	remote := u
	remote.CheckIns = []CheckIn{ci("R1", "01OCT2026", "10:00:00.000")}
	if _, merged, err := s.ApplyRemote(s.encode(remote)); err != nil || !merged {
		t.Fatalf("merged=%v err=%v, want a merge", merged, err)
	}
	got, _ := s.Get(u.UUID)
	if len(got.CheckIns) != 2 || got.CheckIns[1].ULID != "L1" {
		t.Errorf("check-ins = %+v, want R1 then L1", got.CheckIns)
	}
	pending, ok := s.PendingUploads()[u.UUID]
	if !ok {
		t.Fatal("merged record must be queued for upload")
	}
	if p, _ := s.decode(pending); len(p.CheckIns) != 2 {
		t.Errorf("queued record should be the merged one")
	}
}

func TestMarkUploadedKeepsNewerEdit(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.Create("Jane Doe")
	sent := s.PendingUploads()[u.UUID]
	u.PhoneNumber = "edited during upload"
	s.Save(&u, SaveOptions{Remote: true})
	s.MarkUploaded(u.UUID, sent)
	if s.PendingCount() != 1 {
		t.Errorf("an edit made during the upload must stay queued")
	}
}
