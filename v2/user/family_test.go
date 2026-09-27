package user

import "testing"

func members(n int) []Person {
	out := make([]Person, n)
	for i := range out {
		out[i] = Person{Age: 10 + i}
	}
	return out
}

func TestClampFamilyCapsNewRecord(t *testing.T) {
	u := User{FamilySize: 9, FamilyMembers: members(8)}
	ClampFamily(&u, 6, 0)
	if u.FamilySize != 6 {
		t.Errorf("family_size = %d, want 6", u.FamilySize)
	}
	if len(u.FamilyMembers) != 5 {
		t.Errorf("members = %d, want 5 (cap counts the account holder)", len(u.FamilyMembers))
	}
}

func TestClampFamilyLeavesUnderCapAlone(t *testing.T) {
	u := User{FamilySize: 4, FamilyMembers: members(3)}
	ClampFamily(&u, 6, 4)
	if u.FamilySize != 4 || len(u.FamilyMembers) != 3 {
		t.Errorf("a family under the cap must be untouched, got size=%d members=%d", u.FamilySize, len(u.FamilyMembers))
	}
}

// The important one: lowering the setting must not destroy existing records.
func TestClampFamilyNeverShrinksExisting(t *testing.T) {
	u := User{FamilySize: 9, FamilyMembers: members(8)}
	ClampFamily(&u, 6, 9)
	if u.FamilySize != 9 || len(u.FamilyMembers) != 8 {
		t.Errorf("an already-oversized family must keep its members, got size=%d members=%d", u.FamilySize, len(u.FamilyMembers))
	}
}

// ...but it still can't grow further.
func TestClampFamilyRefusesGrowthPastExisting(t *testing.T) {
	u := User{FamilySize: 12, FamilyMembers: members(11)}
	ClampFamily(&u, 6, 9)
	if u.FamilySize != 9 || len(u.FamilyMembers) != 8 {
		t.Errorf("growth past the existing size must be refused, got size=%d members=%d", u.FamilySize, len(u.FamilyMembers))
	}
}

func TestClampFamilyZeroMeansUnlimited(t *testing.T) {
	u := User{FamilySize: 30, FamilyMembers: members(29)}
	ClampFamily(&u, 0, 0)
	if u.FamilySize != 30 || len(u.FamilyMembers) != 29 {
		t.Errorf("0 must mean no limit, got size=%d members=%d", u.FamilySize, len(u.FamilyMembers))
	}
}
