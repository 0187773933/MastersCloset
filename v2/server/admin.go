package server

import (
	fiber "github.com/gofiber/fiber/v2"
)

// These admin JSON endpoints exercise the rewritten user.Store. The richer admin
// HTML pages (search, edit, reports, bulk email/SMS, printing) port in later
// passes; these prove the core operations end-to-end.

// HandleAdminUserGet returns a user record by UUID.
func (s *Server) HandleAdminUserGet(c *fiber.Ctx) error {
	u, ok := s.Users.Get(c.Params("uuid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown user"})
	}
	return c.JSON(u)
}

// HandleAdminUserForce commits a forced (override) check-in.
func (s *Server) HandleAdminUserForce(c *fiber.Ctx) error {
	out, ok := s.Users.CheckInForce(c.Params("uuid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown user"})
	}
	return c.JSON(out)
}

// HandleAdminUserRefill recomputes a user's balance from current config.
func (s *Server) HandleAdminUserRefill(c *fiber.Ctx) error {
	bal, ok := s.Users.Refill(c.Params("uuid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown user"})
	}
	return c.JSON(bal)
}

// HandleAdminUserSimilar lists users similar to the given one.
func (s *Server) HandleAdminUserSimilar(c *fiber.Ctx) error {
	u, ok := s.Users.Get(c.Params("uuid"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown user"})
	}
	return c.JSON(s.Users.FindSimilar(&u))
}

// HandleAdminUserDelete deletes a user and its index entries.
func (s *Server) HandleAdminUserDelete(c *fiber.Ctx) error {
	if err := s.Users.Delete(c.Params("uuid")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}
