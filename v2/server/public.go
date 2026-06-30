package server

import (
	"fmt"

	"github.com/0187773933/MastersCloset/v2/user"
	fiber "github.com/gofiber/fiber/v2"
)

// messageData backs the generic message.html template.
type messageData struct {
	Title    string
	Body     string
	Link     string
	LinkText string
}

const icalLink = "https://calendar.google.com/calendar/ical/masters.closet.5950%40gmail.com/public/basic.ics"

// HandleHome routes to the admin dashboard, the user home, or the public landing
// depending on which cookie (if any) is present.
func (s *Server) HandleHome(c *fiber.Ctx) error {
	if s.isAdmin(c) {
		return c.Redirect("/admin")
	}
	if s.userUUIDFromCookie(c) != "" {
		return s.sendPage(c, "user_home.html")
	}
	return s.sendPage(c, "home.html")
}

// HandleJoin shows the new-user form (or user home if already logged in).
func (s *Server) HandleJoin(c *fiber.Ctx) error {
	if s.userUUIDFromCookie(c) != "" {
		return s.sendPage(c, "user_home.html")
	}
	return s.sendPage(c, "user_new.html")
}

// HandleICal redirects to the public Google calendar feed.
func (s *Server) HandleICal(c *fiber.Ctx) error {
	return c.Redirect(icalLink, fiber.StatusMovedPermanently)
}

// newUserRequest is the JSON body posted by the new-user form.
type newUserRequest struct {
	FirstName    string `json:"first_name"`
	MiddleName   string `json:"middle_name"`
	LastName     string `json:"last_name"`
	EmailAddress string `json:"email_address"`
	PhoneNumber  string `json:"phone_number"`
}

// HandleUserCreate creates a user from the posted form and returns its UUID.
func (s *Server) HandleUserCreate(c *fiber.Ctx) error {
	var req newUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.FirstName == "" && req.LastName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "first or last name required"})
	}
	u, err := s.Users.Create("temp")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	u.Identity.FirstName = req.FirstName
	u.Identity.MiddleName = req.MiddleName
	u.Identity.LastName = req.LastName
	u.EmailAddress = req.EmailAddress
	u.PhoneNumber = req.PhoneNumber
	user.ApplyBalance(&u, s.Cfg.Snapshot().Balance, 1)
	if err := s.Users.Save(&u, user.SaveOptions{Remote: true}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"uuid": u.UUID, "username": u.Username})
}

// HandleLoginFresh performs the QR hand-off: records the first check-in and sets
// the permanent login cookie.
func (s *Server) HandleLoginFresh(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	u, ok := s.Users.Get(uuid)
	if !ok {
		return s.render(c, "message.html", messageData{Title: "Login Failed", Body: "Unknown user."})
	}
	s.Users.CheckIn(u.UUID)
	s.setLongCookie(c, userCookie, u.UUID)
	return s.render(c, "message.html", messageData{
		Title: "Welcome, " + u.NameString, Body: "You're checked in.", Link: "/", LinkText: "Continue",
	})
}

// HandleCheckIn redirects a cookied user to their check-in display.
func (s *Server) HandleCheckIn(c *fiber.Ctx) error {
	uuid := s.userUUIDFromCookie(c)
	if uuid == "" {
		return s.sendPage(c, "user_new.html")
	}
	u, ok := s.Users.Get(uuid)
	if !ok {
		return s.sendPage(c, "user_new.html")
	}
	return c.Redirect(fmt.Sprintf("/user/checkin/display/%s", u.UUID))
}

// HandleCheckInDisplay renders the check-in result for a user.
func (s *Server) HandleCheckInDisplay(c *fiber.Ctx) error {
	out, u, ok := s.Users.CheckInTest(c.Params("uuid"))
	if !ok {
		return s.render(c, "message.html", messageData{Title: "Not Found", Body: "Unknown user."})
	}
	body := "You can check in now."
	if !out.Allowed {
		body = fmt.Sprintf("Please wait %d more day(s) before checking in again.", out.DaysRemaining)
	}
	return s.render(c, "checkin.html", checkinData{
		Name: u.NameString, Allowed: out.Allowed, Message: body, UUID: u.UUID,
	})
}

// checkinData backs checkin.html.
type checkinData struct {
	Name    string
	Allowed bool
	Message string
	UUID    string
}

// HandleCheckInSilent is the passive JSON check-in test.
func (s *Server) HandleCheckInSilent(c *fiber.Ctx) error {
	out, u, ok := s.Users.CheckInTest(c.Params("uuid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown user"})
	}
	return c.JSON(fiber.Map{
		"route": "/user/checkin/silent/:uuid",
		"result": fiber.Map{
			"check_in_possible":      out.Allowed,
			"milliseconds_remaining": out.TimeRemainingMS,
			"days_remaining":         out.DaysRemaining,
			"balance":                u.Balance,
			"name_string":            u.NameString,
			"family_size":            u.FamilySize,
		},
	})
}

// HandleAdminHome renders the admin dashboard.
func (s *Server) HandleAdminHome(c *fiber.Ctx) error {
	return s.sendPage(c, "admin.html")
}
