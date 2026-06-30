package server

import (
	"github.com/0187773933/MastersCloset/v2/config"
	"github.com/0187773933/MastersCloset/v2/printer"
	"github.com/0187773933/MastersCloset/v2/user"
	fiber "github.com/gofiber/fiber/v2"
)

// buildPrintJob maps a user + the shopping ticket they checked in with onto the
// physical ticket. Per-person limits come from live config; the barcode is the
// user's first (often the virtual one minted at check-in).
func buildPrintJob(u user.User, t user.ShoppingTicket, bal config.BalanceConfig) printer.Job {
	per := t.ShoppingFor
	if per < 1 {
		per = 1
	}
	barcode := ""
	if len(u.Barcodes) > 0 {
		barcode = u.Barcodes[0]
	}
	return printer.Job{
		FamilySize:         per,
		TotalClothingItems: per * 6,
		PantsLimit:         bal.General.Bottoms,
		Shoes:              t.Shoes,
		ShoesLimit:         bal.Shoes,
		Accessories:        t.Accessories,
		AccessoriesLimit:   bal.Accessories,
		Seasonal:           t.Seasonals,
		SeasonalLimit:      bal.Seasonals,
		FamilyName:         u.NameString,
		BarcodeNumber:      barcode,
		Language:           u.LangCode(),
		Spanish:            u.LangCode() == "es",
		Boys:               t.Boys,
		Girls:              t.Girls,
		Men:                t.Men,
		Women:              t.Women,
		Guests:             t.Guests,
	}
}

// sendPDF returns rendered ticket bytes as an inline PDF.
func (s *Server) sendPDF(c *fiber.Ctx, job printer.Job) error {
	pdf, err := printer.Render(s.Cfg.Snapshot().Printer, job)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "application/pdf")
	return c.Send(pdf)
}

// HandleLanguages lists the ticket languages the admin UI can choose from.
func (s *Server) HandleLanguages(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"result": printer.Languages()})
}

// HandleTicketStrings returns every language's ticket phrases (the exact strings
// the printer uses) so the live check-in preview mirrors what will print.
func (s *Server) HandleTicketStrings(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"result": printer.LanguageInfos()})
}

// HandlePrint prints a posted print job (manual reprint with custom values).
func (s *Server) HandlePrint(c *fiber.Ctx) error {
	var job printer.Job
	if err := c.BodyParser(&job); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job"})
	}
	printed, err := printer.Print(s.Cfg.Snapshot().Printer, job)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"printed": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"printed": printed})
}

// HandleTicketPDF rebuilds and returns the ticket PDF for a recorded check-in,
// for previewing or re-printing from the browser.
func (s *Server) HandleTicketPDF(c *fiber.Ctx) error {
	ci, ok := s.Users.GetCheckIn(c.Params("ulid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("check-in not found")
	}
	u, ok := s.Users.Get(ci.UUID)
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("user not found")
	}
	return s.sendPDF(c, buildPrintJob(u, ci.Shopping, s.Cfg.Snapshot().Balance))
}
