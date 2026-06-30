package user

import "github.com/0187773933/MastersCloset/v2/config"

// ApplyBalance sets a user's allowance from the configured per-member amounts
// scaled by family size. Single implementation; v1 had this as both a method and
// a standalone function with identical bodies.
func ApplyBalance(u *User, b config.BalanceConfig, familySize int) {
	if familySize < 1 {
		familySize = 1
	}
	u.Balance.General.Total = b.General.Total * familySize
	u.Balance.General.Available = b.General.Total * familySize
	u.Balance.General.Tops.Limit = b.General.Tops * familySize
	u.Balance.General.Tops.Available = b.General.Tops * familySize
	u.Balance.General.Bottoms.Limit = b.General.Bottoms * familySize
	u.Balance.General.Bottoms.Available = b.General.Bottoms * familySize
	u.Balance.General.Dresses.Limit = b.General.Dresses * familySize
	u.Balance.General.Dresses.Available = b.General.Dresses * familySize
	u.Balance.Shoes.Limit = b.Shoes * familySize
	u.Balance.Shoes.Available = b.Shoes * familySize
	u.Balance.Seasonals.Limit = b.Seasonals * familySize
	u.Balance.Seasonals.Available = b.Seasonals * familySize
	u.Balance.Accessories.Limit = b.Accessories * familySize
	u.Balance.Accessories.Available = b.Accessories * familySize
}

// Refill recomputes a user's family size and balance, then persists it.
func (s *Store) Refill(userUUID string) (Balance, bool) {
	u, ok := s.Get(userUUID)
	if !ok {
		return Balance{}, false
	}
	size, _ := u.computeFamilySize()
	ApplyBalance(&u, s.cfg.Snapshot().Balance, size)
	if err := s.Save(&u, SaveOptions{Remote: true}); err != nil {
		return u.Balance, false
	}
	return u.Balance, true
}
