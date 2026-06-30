package user

import (
	json "encoding/json"
	"sort"
	"strconv"
	"time"

	bolt "github.com/boltdb/bolt"
	ulid "github.com/oklog/ulid/v2"
)

// Summary is a lightweight user row for admin list views.
type Summary struct {
	UUID            string `json:"uuid"`
	Username        string `json:"username"`
	NameString      string `json:"name_string"`
	LastCheckInDate string `json:"last_check_in_date"`
	CheckInCount    int    `json:"check_in_count"`
}

// addUsed adds to a running "used" tally.
func addUsed(field *int, amount int) { *field += amount }

// refillSubtract refills the available count from the limit if it has run dry,
// then subtracts what was taken this visit. Ported from v1's _ries.
func refillSubtract(available *int, taken, limit int) {
	if *available < 1 {
		*available = limit
	}
	*available -= taken
}

// mintVirtualBarcode atomically allocates the next high-number virtual barcode.
func (s *Store) mintVirtualBarcode() string {
	bc := ""
	s.db.Update(func(tx *bolt.Tx) error {
		m := tx.Bucket([]byte(bucketMisc))
		idx := 9999999
		if v := m.Get([]byte("virtual-barcode-index")); v != nil {
			if n, err := strconv.Atoi(string(v)); err == nil {
				idx = n
			}
		}
		idx++
		bc = strconv.Itoa(idx)
		return m.Put([]byte("virtual-barcode-index"), []byte(bc))
	})
	return bc
}

// CheckInShopping records an admin "shopping" check-in: it decrements the user's
// per-category balances (refilling from the configured limit when dry), ensures
// the user has a barcode, appends a ULID'd check-in carrying the ticket, and
// persists. This is the real check-in; the spine's CheckIn is only the cooloff
// gate.
func (s *Store) CheckInShopping(userUUID string, t ShoppingTicket) (CheckIn, bool) {
	u, ok := s.Get(userUUID)
	if !ok {
		return CheckIn{}, false
	}
	bal := s.cfg.Snapshot().Balance
	per := t.ShoppingFor
	if per < 1 {
		per = 1
	}

	addUsed(&u.Balance.General.Tops.Used, t.Tops)
	refillSubtract(&u.Balance.General.Tops.Available, t.Tops, bal.General.Tops*per)
	addUsed(&u.Balance.General.Bottoms.Used, t.Bottoms)
	refillSubtract(&u.Balance.General.Bottoms.Available, t.Bottoms, bal.General.Bottoms*per)
	addUsed(&u.Balance.General.Dresses.Used, t.Dresses)
	refillSubtract(&u.Balance.General.Dresses.Available, t.Dresses, bal.General.Dresses*per)
	addUsed(&u.Balance.Shoes.Used, t.Shoes)
	refillSubtract(&u.Balance.Shoes.Available, t.Shoes, bal.Shoes*per)
	addUsed(&u.Balance.Seasonals.Used, t.Seasonals)
	refillSubtract(&u.Balance.Seasonals.Available, t.Seasonals, bal.Seasonals*per)
	addUsed(&u.Balance.Accessories.Used, t.Accessories)
	refillSubtract(&u.Balance.Accessories.Available, t.Accessories, bal.Accessories*per)

	if len(u.Barcodes) == 0 || len(u.Barcodes[0]) < 2 {
		u.Barcodes = append(u.Barcodes, s.mintVirtualBarcode())
	}

	now := s.now()
	ci := CheckIn{
		UUID:     u.UUID,
		Name:     u.NameString,
		ULID:     ulid.Make().String(),
		Date:     dateString(now),
		Time:     timeString(now),
		Type:     "normal",
		Result:   true,
		Shopping: t,
	}
	u.TotalGuestsAdmitted += t.Guests
	u.CheckIns = append(u.CheckIns, ci)

	if err := s.Save(&u, SaveOptions{Remote: true}); err != nil {
		return ci, false
	}
	s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketCheckInIndex)).Put([]byte(ci.ULID), []byte(u.UUID))
	})
	return ci, true
}

// GetCheckIn resolves a check-in by its ULID via the index.
func (s *Store) GetCheckIn(checkInULID string) (CheckIn, bool) {
	var userUUID string
	s.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket([]byte(bucketCheckInIndex)).Get([]byte(checkInULID)); v != nil {
			userUUID = string(v)
		}
		return nil
	})
	if userUUID == "" {
		return CheckIn{}, false
	}
	u, ok := s.Get(userUUID)
	if !ok {
		return CheckIn{}, false
	}
	for _, ci := range u.CheckIns {
		if ci.ULID == checkInULID {
			return ci, true
		}
	}
	return CheckIn{}, false
}

// EditCheckIn replaces a check-in (matched by ULID) and re-saves the owner.
func (s *Store) EditCheckIn(checkInULID string, replacement CheckIn) bool {
	var userUUID string
	s.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket([]byte(bucketCheckInIndex)).Get([]byte(checkInULID)); v != nil {
			userUUID = string(v)
		}
		return nil
	})
	if userUUID == "" {
		return false
	}
	u, ok := s.Get(userUUID)
	if !ok {
		return false
	}
	for i, ci := range u.CheckIns {
		if ci.ULID == checkInULID {
			replacement.ULID = checkInULID
			replacement.UUID = u.UUID
			u.CheckIns[i] = replacement
			return s.Save(&u, SaveOptions{Remote: true}) == nil
		}
	}
	return false
}

// DeleteCheckIn removes a check-in from a user and clears the index entry.
func (s *Store) DeleteCheckIn(userUUID, checkInULID string) bool {
	u, ok := s.Get(userUUID)
	if !ok {
		return false
	}
	kept := u.CheckIns[:0]
	found := false
	for _, ci := range u.CheckIns {
		if ci.ULID == checkInULID {
			found = true
			continue
		}
		kept = append(kept, ci)
	}
	if !found {
		return false
	}
	u.CheckIns = kept
	if err := s.Save(&u, SaveOptions{Remote: true}); err != nil {
		return false
	}
	s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketCheckInIndex)).Delete([]byte(checkInULID))
	})
	return true
}

// DayCheckIns is one collection day with its check-ins (sorted by time).
type DayCheckIns struct {
	Date       string    `json:"date"`
	Count      int       `json:"count"`
	ShoppedFor int       `json:"shopped_for"`
	CheckIns   []CheckIn `json:"check_ins"`
}

// parseDayKey parses a stored "27JUN2026" date for chronological sorting.
func parseDayKey(d string) time.Time {
	t, _ := time.Parse("02Jan2006", d)
	return t
}

// CheckInsByDay returns all check-ins grouped by collection day, newest day
// first, each day's entries sorted by time. Used by the history page and the
// XLSX export.
func (s *Store) CheckInsByDay() []DayCheckIns {
	byDate := map[string][]CheckIn{}
	s.ForEach(func(u User) {
		for _, ci := range u.CheckIns {
			if ci.Name == "" {
				ci.Name = u.NameString
			}
			ci = ci.Normalize(u.FamilySize)
			byDate[ci.Date] = append(byDate[ci.Date], ci)
		}
	})
	dates := make([]string, 0, len(byDate))
	for d := range byDate {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool { return parseDayKey(dates[i]).After(parseDayKey(dates[j])) })

	out := make([]DayCheckIns, 0, len(dates))
	for _, d := range dates {
		cis := byDate[d]
		sort.Slice(cis, func(i, j int) bool { return cis[i].Time < cis[j].Time })
		shopped := 0
		for _, ci := range cis {
			shopped += ci.Shopping.ShoppingFor // normalized: people on the ticket
		}
		out = append(out, DayCheckIns{Date: d, Count: len(cis), ShoppedFor: shopped, CheckIns: cis})
	}
	return out
}

// DaySummaries returns one row per collection day (date, count, shopped-for)
// without the per-check-in arrays — the lightweight feed for the history list.
func (s *Store) DaySummaries() []DayCheckIns {
	days := s.CheckInsByDay()
	for i := range days {
		days[i].CheckIns = nil
	}
	return days
}

// CheckInsByDate returns every (normalized) check-in recorded on the given
// (uppercase) date, sorted by time.
func (s *Store) CheckInsByDate(date string) []CheckIn {
	var out []CheckIn
	s.ForEach(func(u User) {
		for _, ci := range u.CheckIns {
			if ci.Date != date {
				continue
			}
			if ci.Name == "" {
				ci.Name = u.NameString
			}
			out = append(out, ci.Normalize(u.FamilySize))
		}
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Time < out[j].Time })
	return out
}

// CheckInTotals aggregates check-in and shopped-for counts per date.
func (s *Store) CheckInTotals() map[string]map[string]int {
	totals := map[string]map[string]int{}
	s.ForEach(func(u User) {
		for _, ci := range u.CheckIns {
			if totals[ci.Date] == nil {
				totals[ci.Date] = map[string]int{}
			}
			totals[ci.Date]["checkins"]++
			if ci.Shopping.ShoppingFor > 0 {
				totals[ci.Date]["shopped_for"] += ci.Shopping.ShoppingFor
			} else {
				totals[ci.Date]["shopped_for"] += u.FamilySize
			}
		}
	})
	return totals
}

// Summaries returns lightweight rows for every user, for admin list views.
func (s *Store) Summaries() []Summary {
	var out []Summary
	s.ForEach(func(u User) {
		sum := Summary{UUID: u.UUID, Username: u.Username, NameString: u.NameString, CheckInCount: len(u.CheckIns)}
		if len(u.CheckIns) > 0 {
			sum.LastCheckInDate = u.CheckIns[len(u.CheckIns)-1].Date
		}
		out = append(out, sum)
	})
	return out
}

// Update applies a full edited user record (admin edit) and re-saves it. The raw
// JSON body is unmarshaled over a fresh User so callers control the shape.
func (s *Store) Update(raw []byte) (User, error) {
	var u User
	if err := json.Unmarshal(raw, &u); err != nil {
		return u, err
	}
	if err := s.Save(&u, SaveOptions{Remote: true}); err != nil {
		return u, err
	}
	return u, nil
}
