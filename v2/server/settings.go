package server

import (
	"github.com/0187773933/MastersCloset/v2/config"
	"github.com/0187773933/MastersCloset/v2/logger"
	fiber "github.com/gofiber/fiber/v2"
)

// HandleSettingsPage serves the settings panel shell; its JS calls the API below.
func (s *Server) HandleSettingsPage(c *fiber.Ctx) error {
	return s.sendPage(c, "settings.html")
}

// HandleSettingsGet returns the field metadata and current values the panel
// renders. This replaces v1's approach of baking config into client JS by
// rewriting api.js line-by-line at startup.
func (s *Server) HandleSettingsGet(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"fields": config.Fields(),
		"values": s.Cfg.FlatSnapshot(),
	})
}

// HandleSettingsPost applies edited config: it validates, persists to BoltDB,
// and swaps the live snapshot. Safe fields take effect immediately; any
// restart-required fields that changed are reported back to the panel.
func (s *Server) HandleSettingsPost(c *fiber.Ctx) error {
	var edits map[string]interface{}
	if err := c.BodyParser(&edits); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"ok": false, "error": "invalid body"})
	}
	restart, err := s.Cfg.ApplyEdits(edits)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"ok": false, "error": err.Error()})
	}
	logger.GetLogger().Info("settings updated via panel")
	return c.JSON(fiber.Map{"ok": true, "restart_required": restart})
}
