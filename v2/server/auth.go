package server

import (
	"time"

	"github.com/0187773933/MastersCloset/v2/encryption"
	fiber "github.com/gofiber/fiber/v2"
)

const (
	adminCookie = "the-masters-closet-admin"
	userCookie  = "the-masters-closet-user"
)

// isAdmin reports whether the request carries a valid admin cookie. The cookie
// value is a SecretBox-encrypted copy of the configured admin secret message
// (the fiber encryptcookie middleware adds a second transparent layer).
func (s *Server) isAdmin(c *fiber.Ctx) bool {
	raw := c.Cookies(adminCookie)
	if raw == "" {
		return false
	}
	snap := s.Cfg.Snapshot()
	return encryption.SecretBoxDecrypt(snap.BoltDBEncryptionKey, raw) == snap.ServerCookieAdminSecretMessage
}

// userUUIDFromCookie returns the logged-in user's UUID, or "".
func (s *Server) userUUIDFromCookie(c *fiber.Ctx) string {
	raw := c.Cookies(userCookie)
	if raw == "" {
		return ""
	}
	return encryption.SecretBoxDecrypt(s.Cfg.Snapshot().BoltDBEncryptionKey, raw)
}

// requireAdmin is middleware that gates admin routes, redirecting to login.
func (s *Server) requireAdmin(c *fiber.Ctx) error {
	if !s.isAdmin(c) {
		return c.Redirect("/admin/login")
	}
	return c.Next()
}

// setLongCookie stores a 10-year, SecretBox-encrypted cookie.
func (s *Server) setLongCookie(c *fiber.Ctx, name, plaintext string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    encryption.SecretBoxEncrypt(s.Cfg.Snapshot().BoltDBEncryptionKey, plaintext),
		Secure:   true,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Expires:  time.Now().AddDate(10, 0, 0),
	})
}

// HandleAdminLogin renders the login form (GET) or checks credentials (POST).
func (s *Server) HandleAdminLogin(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodGet {
		if s.isAdmin(c) {
			return c.Redirect("/admin")
		}
		return s.sendPage(c, "admin_login.html")
	}
	snap := s.Cfg.Snapshot()
	username := c.FormValue("username")
	password := c.FormValue("password")
	if username != snap.AdminUsername || password != snap.AdminPassword {
		return s.render(c, "message.html", messageData{
			Title: "Login Failed", Body: "Invalid credentials.", Link: "/admin/login", LinkText: "Try again",
		})
	}
	s.setLongCookie(c, adminCookie, snap.ServerCookieAdminSecretMessage)
	return c.Redirect("/admin")
}

// HandleAdminLogout clears the admin cookie.
func (s *Server) HandleAdminLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name: adminCookie, Value: "", Path: "/", Expires: time.Now().Add(-time.Hour), HTTPOnly: true,
	})
	return c.Redirect("/admin/login")
}
