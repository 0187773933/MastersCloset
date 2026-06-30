package server

import (
	json "encoding/json"

	"github.com/0187773933/MastersCloset/v2/logger"
	"github.com/0187773933/MastersCloset/v2/printer"
	"github.com/0187773933/MastersCloset/v2/user"
	fiber "github.com/gofiber/fiber/v2"
)

// AdminCheckInTest returns the cooloff status, the user, and the balance config
// so the check-in screen can render allowances before committing.
func (s *Server) AdminCheckInTest(c *fiber.Ctx) error {
	out, u, ok := s.Users.CheckInTest(c.Params("uuid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown user"})
	}
	return c.JSON(fiber.Map{
		"result":         out,
		"user":           u,
		"balance_config": s.Cfg.Snapshot().Balance,
	})
}

// checkinRequest is the editable print-ticket the admin submits at check-in —
// the fields shown on the ticket preview. Per-person limits default from config
// but can be overridden per visit.
type checkinRequest struct {
	ShoppingFor       int    `json:"shopping_for"`
	ClothingPerPerson int    `json:"clothing_per_person"`
	PantsLimit        int    `json:"pants_limit"`
	ShoesLimit        int    `json:"shoes_limit"`
	AccessoriesLimit  int    `json:"accessories_limit"`
	SeasonalLimit     int    `json:"seasonal_limit"`
	FamilyName        string `json:"family_name"`
	Language          string `json:"language"`
	// Who the family shopped for, from the live "Shopping for" overview. These are
	// recorded for reporting; they do not affect the balance math. Guests are
	// separate — each prints its own ticket but does not change the allowance.
	Boys   int `json:"boys"`
	Girls  int `json:"girls"`
	Men    int `json:"men"`
	Women  int `json:"women"`
	Guests int `json:"guests"`
}

// AdminCheckInShopping records the visit and prints the (edited) ticket exactly
// as previewed. A print failure never fails the check-in; the visit is recorded
// and the ticket can be re-printed from history.
func (s *Server) AdminCheckInShopping(c *fiber.Ctx) error {
	var req checkinRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ticket"})
	}
	// n == 0 is allowed: the family isn't shopping (e.g. only guests are), so no
	// family ticket prints — but guest tickets still do.
	n := req.ShoppingFor
	if n < 0 {
		n = 0
	}
	clothing := req.ClothingPerPerson
	if clothing <= 0 {
		clothing = 6
	}

	// Record the visit (and decrement balance) from the printed allowance. The
	// demographic counts are descriptive only (history/exports); guests are
	// separate and never fold into the family's balance.
	ticket := user.ShoppingTicket{
		ShoppingFor: n,
		Tops:        clothing * n,
		Bottoms:     req.PantsLimit * n,
		Shoes:       req.ShoesLimit * n,
		Seasonals:   req.SeasonalLimit * n,
		Accessories: req.AccessoriesLimit * n,
		Boys:        req.Boys,
		Girls:       req.Girls,
		Men:         req.Men,
		Women:       req.Women,
		Guests:      req.Guests,
	}
	ci, ok := s.Users.CheckInShopping(c.Params("uuid"), ticket)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown user"})
	}

	printed, printErr := false, ""
	guestsPrinted := 0
	if u, found := s.Users.Get(c.Params("uuid")); found {
		barcode := ""
		if len(u.Barcodes) > 0 {
			barcode = u.Barcodes[0]
		}
		name := req.FamilyName
		if name == "" {
			name = u.NameString
		}
		lang := req.Language
		if lang == "" {
			lang = u.LangCode()
		}
		job := printer.Job{
			FamilySize:         n,
			TotalClothingItems: clothing * n,
			ClothingPerPerson:  clothing,
			PantsLimit:         req.PantsLimit,
			ShoesLimit:         req.ShoesLimit,
			AccessoriesLimit:   req.AccessoriesLimit,
			SeasonalLimit:      req.SeasonalLimit,
			FamilyName:         name,
			BarcodeNumber:      barcode,
			Language:           lang,
		}
		// Skip the family ticket entirely when nobody in the family is shopping
		// (n == 0); the guests below still print on their own.
		if n >= 1 {
			var perr error
			if printed, perr = printer.Print(s.Cfg.Snapshot().Printer, job); perr != nil {
				printErr = perr.Error()
				logger.GetLogger().Info("check-in ticket print failed: " + printErr)
			}
		}

		// Each guest gets its own full ticket — a 1-person allowance titled
		// "Guest ( N )" carrying the same barcode, so it scans to this user.
		for g := 1; g <= req.Guests; g++ {
			guestJob := job
			guestJob.GuestNumber = g
			guestJob.FamilySize = 1
			guestJob.TotalClothingItems = clothing
			gPrinted, gErr := printer.Print(s.Cfg.Snapshot().Printer, guestJob)
			if gErr != nil {
				if printErr == "" {
					printErr = gErr.Error()
				}
				logger.GetLogger().Info("guest ticket print failed: " + gErr.Error())
				continue
			}
			if gPrinted {
				guestsPrinted++
			}
		}
	}
	return c.JSON(fiber.Map{"result": true, "check_in": ci, "printed": printed, "print_error": printErr, "guests_printed": guestsPrinted})
}

// AdminCheckInTotals aggregates check-in counts per date.
func (s *Server) AdminCheckInTotals(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"result": s.Users.CheckInTotals()})
}

// AdminCheckInsByDate lists every check-in on a given (uppercase) date.
func (s *Server) AdminCheckInsByDate(c *fiber.Ctx) error {
	date := c.Params("date")
	return c.JSON(fiber.Map{"date": date, "result": s.Users.CheckInsByDate(date)})
}

// AdminGetCheckIn returns one check-in by ULID.
func (s *Server) AdminGetCheckIn(c *fiber.Ctx) error {
	ci, ok := s.Users.GetCheckIn(c.Params("ulid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown check-in"})
	}
	return c.JSON(fiber.Map{"result": ci})
}

// AdminEditCheckIn replaces a check-in (matched by ULID) with the posted record.
func (s *Server) AdminEditCheckIn(c *fiber.Ctx) error {
	var replacement user.CheckIn
	if err := json.Unmarshal(c.Body(), &replacement); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid check-in"})
	}
	if !s.Users.EditCheckIn(c.Params("ulid"), replacement) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "check-in not found"})
	}
	return c.JSON(fiber.Map{"result": true})
}

// AdminDeleteCheckIn removes a check-in from a user.
func (s *Server) AdminDeleteCheckIn(c *fiber.Ctx) error {
	if !s.Users.DeleteCheckIn(c.Params("uuid"), c.Params("ulid")) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "check-in not found"})
	}
	return c.JSON(fiber.Map{"result": "deleted"})
}
