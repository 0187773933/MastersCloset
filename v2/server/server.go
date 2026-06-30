// Package server is the v2 HTTP layer. The Server struct is the "monolith var"
// the README asked for: it carries the config Manager, db, search index, and the
// user Store, and every handler is a method that reads config live via
// s.Cfg.Snapshot(). All assets are embedded (see assets.go), so no handler
// references the filesystem by string literal.
package server

import (
	"bytes"
	"fmt"

	"github.com/0187773933/MastersCloset/v2/config"
	"github.com/0187773933/MastersCloset/v2/logger"
	"github.com/0187773933/MastersCloset/v2/user"
	bleve "github.com/blevesearch/bleve/v2"
	bolt "github.com/boltdb/bolt"
	fiber "github.com/gofiber/fiber/v2"
	fiber_cors "github.com/gofiber/fiber/v2/middleware/cors"
	fiber_cookie "github.com/gofiber/fiber/v2/middleware/encryptcookie"
)

// Server bundles every dependency the HTTP handlers need.
type Server struct {
	App   *fiber.App
	Cfg   *config.Manager
	DB    *bolt.DB
	Index bleve.Index
	Users *user.Store
}

// New builds a configured (but not yet listening) server.
func New(cfg *config.Manager, db *bolt.DB, index bleve.Index, users *user.Store) *Server {
	s := &Server{
		App:   fiber.New(fiber.Config{DisableStartupMessage: true}),
		Cfg:   cfg,
		DB:    db,
		Index: index,
		Users: users,
	}

	snap := cfg.Snapshot()

	s.App.Use(func(c *fiber.Ctx) error {
		ip := c.Get("x-forwarded-for")
		if ip == "" {
			ip = c.IP()
		}
		logger.GetLogger().Debug(fmt.Sprintf("%s === %s === %s", ip, c.Method(), c.Path()))
		return c.Next()
	})

	s.App.Use(fiber_cookie.New(fiber_cookie.Config{Key: snap.ServerCookieSecret}))

	origins := fmt.Sprintf("%s, %s", snap.ServerBaseURL, snap.ServerLiveURL)
	s.App.Use(fiber_cors.New(fiber_cors.Config{
		AllowOrigins:     origins,
		AllowHeaders:     "Origin, Content-Type, Accept, key",
		AllowCredentials: true,
	}))

	s.registerRoutes()
	return s
}

// Start binds and serves. Blocks until shutdown.
func (s *Server) Start() error {
	snap := s.Cfg.Snapshot()
	log := logger.GetLogger()
	log.Info(fmt.Sprintf("Listening on http://localhost:%s", snap.ServerPort))
	log.Info(fmt.Sprintf("Admin login @ http://localhost:%s/admin/login", snap.ServerPort))
	log.Info(fmt.Sprintf("Settings panel @ http://localhost:%s/admin/settings", snap.ServerPort))
	return s.App.Listen(":" + snap.ServerPort)
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown() error { return s.App.Shutdown() }

// render executes a parsed template into an HTML response.
func (s *Server) render(c *fiber.Ctx, name string, data interface{}) error {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("template error: " + err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}

// sendPage serves a raw embedded html/ page.
func (s *Server) sendPage(c *fiber.Ctx, name string) error {
	b, err := page(name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("page not found: " + name)
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(b)
}
