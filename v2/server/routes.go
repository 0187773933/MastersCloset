package server

import (
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
)

// publicLimiter throttles public endpoints (ported from v1).
var publicLimiter = rate_limiter.New(rate_limiter.Config{
	Max:        30,
	Expiration: time.Second,
	KeyGenerator: func(c *fiber.Ctx) string {
		if ip := c.Get("x-forwarded-for"); ip != "" {
			return ip
		}
		return c.IP()
	},
	LimitReached: func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(`<html><h1>loading ...</h1><script>setTimeout(function(){ window.location.reload(1); }, 6000);</script></html>`)
	},
})

// registerRoutes wires every route. Static assets are served from the embedded
// filesystem, so there is no on-disk document root.
func (s *Server) registerRoutes() {
	// Embedded static assets at /static. No long-lived caching: the assets are
	// embedded in the binary, so a rebuild should be picked up immediately
	// rather than served stale from the browser cache for an hour.
	s.App.Use("/static", func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache")
		return c.Next()
	}, filesystem.New(filesystem.Config{Root: staticHTTPFS()}))
	s.App.Get("/favicon.ico", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "image/x-icon")
		b, err := staticFS.ReadFile("static/favicon.ico")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.Send(b)
	})

	// Public flow.
	s.App.Get("/", publicLimiter, s.HandleHome)
	s.App.Get("/join", publicLimiter, s.HandleJoin)
	s.App.Get("/checkin", publicLimiter, s.HandleCheckIn)
	s.App.Get("/ical", publicLimiter, s.HandleICal)

	api := s.App.Group("/user")
	api.Post("/api/new", publicLimiter, s.HandleUserCreate)
	api.Get("/login/fresh/:uuid", publicLimiter, s.HandleLoginFresh)
	api.Get("/checkin", publicLimiter, s.HandleCheckIn)
	api.Get("/checkin/display/:uuid", publicLimiter, s.HandleCheckInDisplay)
	api.Get("/checkin/silent/:uuid", publicLimiter, s.HandleCheckInSilent)

	// Admin auth.
	s.App.Get("/admin/login", s.HandleAdminLogin)
	s.App.Post("/admin/login", s.HandleAdminLogin)
	s.App.Get("/admin/logout", s.HandleAdminLogout)

	// Admin (gated).
	admin := s.App.Group("/admin", s.requireAdmin)
	admin.Get("/", s.HandleAdminHome)

	// Admin HTML pages. The `/:uuid` variants are deep-links to a preselected
	// user (the page reads the uuid from the path), so they survive a reload and
	// can be shared/bookmarked.
	admin.Get("/users", func(c *fiber.Ctx) error { return s.sendPage(c, "admin_users.html") })
	admin.Get("/users/:uuid", func(c *fiber.Ctx) error { return s.sendPage(c, "admin_users.html") })
	admin.Get("/user/new", func(c *fiber.Ctx) error { return s.sendPage(c, "admin_user_new.html") })
	admin.Get("/checkin", func(c *fiber.Ctx) error { return s.sendPage(c, "admin_checkin.html") })
	admin.Get("/checkin/:uuid", func(c *fiber.Ctx) error { return s.sendPage(c, "admin_checkin.html") })

	// Settings.
	admin.Get("/settings", s.HandleSettingsPage)
	admin.Get("/settings/api", s.HandleSettingsGet)
	admin.Post("/settings/api", s.HandleSettingsPost)

	// User management API.
	admin.Get("/user/all", s.AdminUserList)
	admin.Get("/user/get/:uuid", s.HandleAdminUserGet)
	admin.Get("/user/barcode/:barcode", s.AdminUserByBarcode)
	admin.Get("/user/search/fuzzy/:query", s.AdminUserSearchFuzzy)
	admin.Get("/user/search/:name", s.AdminUserSearchExact)
	admin.Get("/user/similar/:uuid", s.HandleAdminUserSimilar)
	admin.Get("/user/handoff/:uuid/qr.png", s.AdminHandoffQR)
	admin.Post("/user/new", s.AdminUserCreate)
	admin.Post("/user/edit", s.AdminUserEdit)
	admin.Delete("/user/:uuid", s.HandleAdminUserDelete)

	// Check-in API.
	admin.Get("/user/checkin/test/:uuid", s.AdminCheckInTest)
	admin.Post("/user/checkin/force/:uuid", s.HandleAdminUserForce)
	admin.Post("/user/checkin/:uuid", s.AdminCheckInShopping)
	admin.Post("/user/refill/:uuid", s.HandleAdminUserRefill)
	admin.Get("/checkins", func(c *fiber.Ctx) error { return s.sendPage(c, "admin_checkins.html") })
	admin.Get("/checkins/days", s.AdminCheckInDaySummaries)
	admin.Get("/checkins/totals", s.AdminCheckInTotals)
	admin.Get("/checkins/export.xlsx", s.AdminCheckInsXLSX)
	admin.Get("/checkins/day/:date", func(c *fiber.Ctx) error { return s.sendPage(c, "admin_checkin_day.html") })
	admin.Get("/checkins/day/:date/export.xlsx", s.AdminCheckInDayXLSX)
	admin.Get("/checkins/date/:date", s.AdminCheckInsByDate)
	admin.Get("/checkin/:ulid/ticket.pdf", s.HandleTicketPDF)
	// Single check-in record ops keyed by ULID. They live two segments deep
	// (mirroring Delete) so the one-segment `/checkin/:uuid` page route above
	// stays unambiguous. The handlers only read :ulid; :uuid is ignored.
	admin.Get("/checkin/:uuid/:ulid", s.AdminGetCheckIn)
	admin.Post("/checkin/:uuid/:ulid", s.AdminEditCheckIn)
	admin.Delete("/checkin/:uuid/:ulid", s.AdminDeleteCheckIn)

	// Printing (reprint a recorded ticket, or print a custom job).
	admin.Get("/languages", s.HandleLanguages)
	admin.Get("/ticket/strings", s.HandleTicketStrings)
	admin.Post("/print", s.HandlePrint)

	// Fallback.
	s.App.Get("/*", publicLimiter, func(c *fiber.Ctx) error { return c.Redirect("/") })
}
