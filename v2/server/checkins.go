package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/0187773933/MastersCloset/v2/user"
	fiber "github.com/gofiber/fiber/v2"
	excelize "github.com/xuri/excelize/v2"
)

// AdminCheckInDaySummaries returns one row per collection day (date, count,
// shopped-for) — the lightweight feed the history list renders.
func (s *Server) AdminCheckInDaySummaries(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"result": s.Users.DaySummaries()})
}

// dayRowHeaders are the per-check-in columns shared by both exports.
var dayRowHeaders = []string{"Name", "Time", "Type", "Shopping For", "Tops", "Bottoms", "Dresses",
	"Shoes", "Seasonals", "Accessories", "Boys", "Girls", "Men", "Women", "Guests", "Check-In ID"}

// sheetName makes a date safe and unique as an Excel sheet name (≤31 chars).
func sheetName(date string, used map[string]bool) string {
	r := strings.NewReplacer(`:`, "-", `\`, "-", `/`, "-", `?`, "", `*`, "", `[`, "(", `]`, ")")
	name := r.Replace(date)
	if name == "" {
		name = "Day"
	}
	if len(name) > 31 {
		name = name[:31]
	}
	base, i := name, 2
	for used[name] {
		suffix := fmt.Sprintf(" (%d)", i)
		if len(base) > 31-len(suffix) {
			name = base[:31-len(suffix)] + suffix
		} else {
			name = base + suffix
		}
		i++
	}
	used[name] = true
	return name
}

// writeTotalsSheet writes the grand-totals sheet from per-day summaries.
func writeTotalsSheet(f *excelize.File, sheet string, days []user.DayCheckIns) {
	f.SetCellValue(sheet, "A1", "Master's Closet — Check-In Snapshot")
	f.SetCellValue(sheet, "A2", "Generated")
	f.SetCellValue(sheet, "B2", time.Now().Format("2006-01-02 15:04"))
	f.SetCellValue(sheet, "A4", "Date")
	f.SetCellValue(sheet, "B4", "Check-Ins")
	f.SetCellValue(sheet, "C4", "People Shopped For")
	row, gc, gs := 5, 0, 0
	for _, d := range days {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), d.Date)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), d.Count)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), d.ShoppedFor)
		gc += d.Count
		gs += d.ShoppedFor
		row++
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "GRAND TOTAL")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), gc)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), gs)
}

// writeDaySheet writes one collection day's check-ins onto a sheet.
func writeDaySheet(f *excelize.File, sheet string, checkins []user.CheckIn) {
	for i, h := range dayRowHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for r, ci := range checkins {
		sh := ci.Shopping
		vals := []interface{}{ci.Name, ci.Time, ci.Type, sh.ShoppingFor, sh.Tops, sh.Bottoms, sh.Dresses,
			sh.Shoes, sh.Seasonals, sh.Accessories, sh.Boys, sh.Girls, sh.Men, sh.Women, sh.Guests, ci.ULID}
		for i, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(i+1, r+2)
			f.SetCellValue(sheet, cell, v)
		}
	}
}

// stream writes the workbook as an .xlsx attachment.
func stream(c *fiber.Ctx, f *excelize.File, filename string) error {
	buf, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.Send(buf.Bytes())
}

// AdminCheckInsXLSX streams the full snapshot: first sheet grand totals, then one
// sheet per collection day.
func (s *Server) AdminCheckInsXLSX(c *fiber.Ctx) error {
	days := s.Users.CheckInsByDay()
	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", "Totals")
	writeTotalsSheet(f, "Totals", days)
	used := map[string]bool{"Totals": true}
	for _, d := range days {
		name := sheetName(d.Date, used)
		f.NewSheet(name)
		writeDaySheet(f, name, d.CheckIns)
	}
	f.SetActiveSheet(0)
	return stream(c, f, fmt.Sprintf("checkins_%s.xlsx", time.Now().Format("20060102")))
}

// AdminCheckInDayXLSX streams a one-day report: a totals sheet plus that day's
// check-ins.
func (s *Server) AdminCheckInDayXLSX(c *fiber.Ctx) error {
	date := c.Params("date")
	checkins := s.Users.CheckInsByDate(date)
	shopped := 0
	for _, ci := range checkins {
		shopped += ci.Shopping.ShoppingFor
	}
	summary := []user.DayCheckIns{{Date: date, Count: len(checkins), ShoppedFor: shopped}}

	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", "Totals")
	writeTotalsSheet(f, "Totals", summary)
	used := map[string]bool{"Totals": true}
	name := sheetName(date, used)
	f.NewSheet(name)
	writeDaySheet(f, name, checkins)
	f.SetActiveSheet(0)
	return stream(c, f, fmt.Sprintf("checkins_%s.xlsx", date))
}
