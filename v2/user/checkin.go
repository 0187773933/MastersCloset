package user

import (
	"time"
)

// Outcome is the result of evaluating whether a user may check in. v1 returned
// these values as loose multi-returns from four near-identical functions.
type Outcome struct {
	Allowed         bool   `json:"allowed"`
	TimeRemainingMS int    `json:"time_remaining_ms"` // 0 when allowed
	DaysRemaining   int    `json:"days_remaining"`
	Date            string `json:"date"`
	Time            string `json:"time"`
	FirstEver       bool   `json:"first_ever"`
}

// Evaluate is THE check-in cooloff calculation — the single source of truth that
// both the passive test and the committing paths use. Negative remaining means
// the user has waited long enough.
func Evaluate(u *User, now time.Time, cooloffDays int) Outcome {
	out := Outcome{Date: dateString(now), Time: timeString(now)}
	if len(u.CheckIns) == 0 {
		out.Allowed = true
		out.FirstEver = true
		return out
	}
	last := u.CheckIns[len(u.CheckIns)-1]
	lastDate, _ := time.ParseInLocation("02Jan2006", last.Date, now.Location())
	cooloff := time.Duration(cooloffDays) * 24 * time.Hour
	remaining := cooloff - now.Sub(lastDate)
	if remaining < 0 {
		out.Allowed = true
		return out
	}
	out.Allowed = false
	out.DaysRemaining = int(remaining.Hours() / 24)
	out.TimeRemainingMS = int(remaining.Milliseconds())
	return out
}

// CheckInTest evaluates a check-in without recording it (the "silent" path).
func (s *Store) CheckInTest(userUUID string) (Outcome, User, bool) {
	u, ok := s.Get(userUUID)
	if !ok {
		return Outcome{}, u, false
	}
	out := Evaluate(&u, s.now(), s.cfg.Snapshot().CheckInCoolOffDays)
	u.AllowedToCheckIn = out.Allowed
	if out.Allowed {
		u.TimeRemaining = 0
	} else {
		u.TimeRemaining = out.TimeRemainingMS
	}
	return out, u, true
}

// CheckIn records a check-in if allowed, or a failed attempt otherwise.
func (s *Store) CheckIn(userUUID string) (Outcome, bool) {
	u, ok := s.Get(userUUID)
	if !ok {
		return Outcome{}, false
	}
	out := Evaluate(&u, s.now(), s.cfg.Snapshot().CheckInCoolOffDays)
	entry := CheckIn{UUID: u.UUID, Date: out.Date, Time: out.Time, Result: out.Allowed}
	if out.Allowed {
		if out.FirstEver {
			entry.Type = "first"
		} else {
			entry.Type = "new"
		}
		u.CheckIns = append(u.CheckIns, entry)
	} else {
		entry.Type = "failed"
		entry.TimeRemaining = out.TimeRemainingMS
		u.FailedCheckIns = append(u.FailedCheckIns, entry)
	}
	u.AllowedToCheckIn = out.Allowed
	u.TimeRemaining = out.TimeRemainingMS
	if err := s.Save(&u, SaveOptions{Remote: true}); err != nil {
		return out, false
	}
	return out, true
}

// CheckInForce records a check-in unconditionally (admin override).
func (s *Store) CheckInForce(userUUID string) (Outcome, bool) {
	u, ok := s.Get(userUUID)
	if !ok {
		return Outcome{}, false
	}
	out := Evaluate(&u, s.now(), s.cfg.Snapshot().CheckInCoolOffDays)
	out.Allowed = true
	out.TimeRemainingMS = 0
	u.CheckIns = append(u.CheckIns, CheckIn{
		UUID:   u.UUID,
		Date:   out.Date,
		Time:   out.Time,
		Type:   "forced",
		Result: true,
	})
	u.AllowedToCheckIn = true
	u.TimeRemaining = 0
	if err := s.Save(&u, SaveOptions{Remote: true}); err != nil {
		return out, false
	}
	return out, true
}
