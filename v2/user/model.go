// Package user is the v2 rewrite of the user domain. v1's user.go mixed data,
// persistence, and four copies of the check-in math into one 900-line file with
// a User struct that embedded *bolt.DB and *config. v2 splits it: this file is
// pure data, store.go owns persistence, checkin.go owns the (single) cooloff
// calculation, balance.go and similar.go own their concerns.
package user

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ShoppingTicket captures what a family shopped for during an admin check-in.
// v1 buried these on printer.PrintJob; here they live on the check-in record so
// the data survives without a printer dependency.
type ShoppingTicket struct {
	ShoppingFor int `json:"shopping_for"` // number of people shopped for
	Tops        int `json:"tops"`
	Bottoms     int `json:"bottoms"`
	Dresses     int `json:"dresses"`
	Shoes       int `json:"shoes"`
	Seasonals   int `json:"seasonals"`
	Accessories int `json:"accessories"`
	Boys        int `json:"boys"`
	Girls       int `json:"girls"`
	Men         int `json:"men"`
	Women       int `json:"women"`
	Guests      int `json:"guests"`
}

// LegacyPrintJob preserves the v1 ticket fields. v1 stored a check-in's shopping
// detail (notably the family size that was on the ticket) under "print_job". v2
// uses Shopping going forward, but reads this so historical check-ins still
// report the correct number of people shopped for.
type LegacyPrintJob struct {
	FamilySize int `json:"family_size"`
	Boys       int `json:"boys"`
	Girls      int `json:"girls"`
	Men        int `json:"men"`
	Women      int `json:"women"`
	Guests     int `json:"guests"`
}

// CheckIn records one (successful or failed) check-in attempt. New records carry
// the Shopping ticket; v1 records carry PrintJob (preserved on re-save).
type CheckIn struct {
	UUID          string          `json:"uuid"`
	Name          string          `json:"name"`
	ULID          string          `json:"ULID"`
	Date          string          `json:"date"`
	Time          string          `json:"time"`
	Type          string          `json:"type"`
	Result        bool            `json:"result"`
	TimeRemaining int             `json:"time_remaining"`
	Shopping      ShoppingTicket  `json:"shopping"`
	PrintJob      *LegacyPrintJob `json:"print_job,omitempty"`
}

// Normalize returns a display copy whose Shopping ticket is filled in for v1
// records: ShoppingFor becomes the people on the ticket (legacy family size, or
// the user's current family size as a last resort) and demographic counts are
// pulled from the legacy print job. It never mutates stored data.
func (ci CheckIn) Normalize(userFamilySize int) CheckIn {
	if ci.Shopping.ShoppingFor <= 0 {
		switch {
		case ci.PrintJob != nil && ci.PrintJob.FamilySize > 0:
			ci.Shopping.ShoppingFor = ci.PrintJob.FamilySize
		case userFamilySize > 0:
			ci.Shopping.ShoppingFor = userFamilySize
		default:
			ci.Shopping.ShoppingFor = 1
		}
	}
	if ci.PrintJob != nil {
		if ci.Shopping.Boys == 0 {
			ci.Shopping.Boys = ci.PrintJob.Boys
		}
		if ci.Shopping.Girls == 0 {
			ci.Shopping.Girls = ci.PrintJob.Girls
		}
		if ci.Shopping.Men == 0 {
			ci.Shopping.Men = ci.PrintJob.Men
		}
		if ci.Shopping.Women == 0 {
			ci.Shopping.Women = ci.PrintJob.Women
		}
		if ci.Shopping.Guests == 0 {
			ci.Shopping.Guests = ci.PrintJob.Guests
		}
	}
	return ci
}

// BalanceItem is one allowance line (e.g. tops): a limit and what's left.
type BalanceItem struct {
	Available int `json:"available"`
	Limit     int `json:"limit"`
	Used      int `json:"used"`
}

// GeneralClothes is the general-clothing sub-balance.
type GeneralClothes struct {
	Total     int         `json:"total"`
	Available int         `json:"available"`
	Tops      BalanceItem `json:"tops"`
	Bottoms   BalanceItem `json:"bottoms"`
	Dresses   BalanceItem `json:"dresses"`
}

// Balance is a user's current per-period allowance state.
type Balance struct {
	General     GeneralClothes `json:"general"`
	Shoes       BalanceItem    `json:"shoes"`
	Seasonals   BalanceItem    `json:"seasonals"`
	Accessories BalanceItem    `json:"accessories"`
}

// DateOfBirth is a loosely-typed birthday (month as a name string, as in v1).
type DateOfBirth struct {
	Month string `json:"month"`
	Day   int    `json:"day"`
	Year  int    `json:"year"`
}

// Address is a US mailing address.
type Address struct {
	StreetNumber string `json:"street_number"`
	StreetName   string `json:"street_name"`
	AddressTwo   string `json:"address_two"`
	City         string `json:"city"`
	State        string `json:"state"`
	ZipCode      string `json:"zipcode"`
}

// Person is one identity (the account holder or a family member/alias).
type Person struct {
	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	MiddleName  string      `json:"middle_name"`
	Address     Address     `json:"address"`
	DateOfBirth DateOfBirth `json:"date_of_birth"`
	Age         int         `json:"age"`
	Sex         string      `json:"sex"`
	Height      string      `json:"height"`
	EyeColor    string      `json:"eye_color"`
	Spouse      bool        `json:"spouse"`
}

// User is the persisted account record. It is pure data — no db or config
// handles live here; the Store supplies those.
type User struct {
	Verified            bool      `json:"verified"`
	Username            string    `json:"username"`
	NameString          string    `json:"name_string"`
	SearchString        string    `json:"search_string"`
	UUID                string    `json:"uuid"`
	ULID                string    `json:"ulid"`
	Barcodes            []string  `json:"barcodes"`
	EmailAddress        string    `json:"email_address"`
	PhoneNumber         string    `json:"phone_number"`
	Identity            Person    `json:"identity"`
	AuthorizedAliases   []Person  `json:"authorized_aliases"`
	FamilySize          int       `json:"family_size"`
	FamilyMembers       []Person  `json:"family_members"`
	CreatedDate         string    `json:"created_date"`
	CreatedTime         string    `json:"created_time"`
	ModifiedDate        string    `json:"modified_date"`
	ModifiedTime        string    `json:"modified_time"`
	CheckIns            []CheckIn `json:"check_ins"`
	FailedCheckIns      []CheckIn `json:"failed_check_ins"`
	TotalGuestsAdmitted int       `json:"total_guests_admitted"`
	Balance             Balance   `json:"balance"`
	TimeRemaining       int       `json:"time_remaining"`
	AllowedToCheckIn    bool      `json:"allowed_to_checkin"`
	Spanish             bool      `json:"spanish"`  // legacy; kept for back-compat
	Language            string    `json:"language"` // preferred language code, e.g. "en", "es"
	SimilarUsers        []User    `json:"similar_users"`
}

// LangCode resolves the user's effective language: the explicit Language field,
// falling back to the legacy Spanish bool, then English. This lets existing
// records (which only have `spanish`) keep working.
func (u *User) LangCode() string {
	if u.Language != "" {
		return u.Language
	}
	if u.Spanish {
		return "es"
	}
	return "en"
}

var titleCaser = cases.Title(language.AmericanEnglish)

// titleCase replaces v1's deprecated strings.Title with the modern caser.
func titleCase(s string) string {
	return titleCaser.String(strings.ToLower(strings.TrimSpace(s)))
}

// FormatUsername normalizes identity name casing and derives the username,
// display name, and lowercase search string. This is the single canonical
// implementation; v1 had it duplicated as both a method and a free function.
func FormatUsername(u *User) {
	var parts []string
	if u.Identity.FirstName != "" {
		u.Identity.FirstName = titleCase(u.Identity.FirstName)
		parts = append(parts, u.Identity.FirstName)
	}
	if u.Identity.MiddleName != "" {
		u.Identity.MiddleName = titleCase(u.Identity.MiddleName)
		parts = append(parts, u.Identity.MiddleName)
	}
	if u.Identity.LastName != "" {
		u.Identity.LastName = titleCase(u.Identity.LastName)
		parts = append(parts, u.Identity.LastName)
	}
	if len(parts) > 0 {
		u.Username = strings.Join(parts, "-")
		u.NameString = strings.Join(parts, " ")
	}
	u.SearchString = strings.ToLower(u.NameString)
}

// FamilySize returns members + the account holder, keeping the cached field in
// sync. Returns the computed size and whether it changed.
func (u *User) computeFamilySize() (size int, changed bool) {
	size = len(u.FamilyMembers) + 1
	if u.FamilySize != size {
		u.FamilySize = size
		changed = true
	}
	return
}
