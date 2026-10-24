package server

import (
	fiber "github.com/gofiber/fiber/v2"
)

// HandleSyncStatus reports the remote db sync state for the dashboard.
func (s *Server) HandleSyncStatus(c *fiber.Ctx) error {
	return c.JSON(s.Sync.Status())
}

// HandleSyncNow runs a pull/push round immediately and returns the result.
func (s *Server) HandleSyncNow(c *fiber.Ctx) error {
	return c.JSON(s.Sync.SyncNow())
}
