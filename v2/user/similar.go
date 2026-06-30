package user

import (
	distance "github.com/hbollon/go-edlib"
)

// SimilarReport flags the dimensions on which two users look like duplicates.
type SimilarReport struct {
	IsSimilar bool `json:"is_similar"`
	Name      bool `json:"name"`
	Email     bool `json:"email"`
	Phone     bool `json:"phone"`
	Address   bool `json:"address"`
	Birthday  bool `json:"birthday"`
	Barcode   bool `json:"barcode"`
	User      User `json:"user"`
}

// Compare builds a similarity report between two users using a Levenshtein
// threshold for fuzzy fields.
func Compare(a, b *User, threshold int) SimilarReport {
	r := SimilarReport{User: *b}
	r.Name = similarByName(a, b, threshold)
	r.Email = similarByEmail(a, b, threshold)
	r.Phone = similarByPhone(a, b)
	r.Address = similarByAddress(a, b, threshold)
	r.Birthday = similarByBirthday(a, b)
	r.Barcode = similarByBarcode(a, b)
	r.IsSimilar = r.Name || r.Email || r.Phone || r.Address || r.Birthday || r.Barcode
	return r
}

// FindSimilar scans every user and returns those similar to the given one
// (excluding itself).
func (s *Store) FindSimilar(u *User) []User {
	threshold := s.cfg.Snapshot().LevenshteinDistanceThreshold
	var out []User
	s.ForEach(func(other User) {
		if other.UUID == u.UUID {
			return
		}
		if Compare(u, &other, threshold).IsSimilar {
			out = append(out, other)
		}
	})
	return out
}

func within(a, b string, threshold int) bool {
	if a == "" || b == "" {
		return false
	}
	return distance.LevenshteinDistance(a, b) < threshold
}

func similarByName(a, b *User, threshold int) bool {
	return within(a.Identity.FirstName, b.Identity.FirstName, threshold) &&
		within(a.Identity.LastName, b.Identity.LastName, threshold)
}

// similarByEmail compares the email addresses. v1's version compared last names
// here by mistake — fixed.
func similarByEmail(a, b *User, threshold int) bool {
	return within(a.EmailAddress, b.EmailAddress, threshold)
}

func similarByPhone(a, b *User) bool {
	return a.PhoneNumber != "" && a.PhoneNumber == b.PhoneNumber
}

func similarByAddress(a, b *User, threshold int) bool {
	return within(a.Identity.Address.StreetNumber, b.Identity.Address.StreetNumber, threshold) &&
		within(a.Identity.Address.StreetName, b.Identity.Address.StreetName, threshold)
}

func similarByBirthday(a, b *User) bool {
	d1, d2 := a.Identity.DateOfBirth, b.Identity.DateOfBirth
	if d1.Year == 0 || d1.Month == "" || d1.Day == 0 {
		return false
	}
	return d1.Year == d2.Year && d1.Month == d2.Month && d1.Day == d2.Day
}

func similarByBarcode(a, b *User) bool {
	for _, x := range a.Barcodes {
		if x == "" {
			continue
		}
		for _, y := range b.Barcodes {
			if y != "" && x == y {
				return true
			}
		}
	}
	return false
}
