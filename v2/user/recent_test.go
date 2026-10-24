package user

import "testing"

func TestRecentCheckInsNewestFirstSkipsFailedAndLimits(t *testing.T) {
	s := newTestStore(t)
	add := func(name string, cis ...CheckIn) {
		u, err := s.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		u.NameString = name
		u.CheckIns = cis
		if err := s.Save(&u, SaveOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	failed := ci("F", "25OCT2026", "09:00:00.000")
	failed.Result = false

	add("old", ci("A", "01SEP2026", "10:00:00.000"))
	// Their failed attempt is the newest of anyone's, but must not count.
	add("mid", ci("B", "10OCT2026", "10:00:00.000"), failed)
	add("new", ci("C", "20OCT2026", "08:00:00.000"))
	add("never", failed)

	got := s.RecentCheckIns(2)
	if len(got) != 2 || got[0].NameString != "new" || got[1].NameString != "mid" {
		t.Fatalf("got %+v, want [new mid]", got)
	}
	if got[1].LastCheckInDate != "10OCT2026" || got[1].LastCheckInTime != "10:00:00.000" {
		t.Errorf("mid last check-in = %s %s, want its successful one", got[1].LastCheckInDate, got[1].LastCheckInTime)
	}
	if all := s.RecentCheckIns(25); len(all) != 3 {
		t.Errorf("got %d users, want 3 (never-checked-in excluded)", len(all))
	}
}
