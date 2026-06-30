package server

import (
	"bytes"
	json "encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/0187773933/MastersCloset/v2/logger"
	"github.com/0187773933/MastersCloset/v2/user"
	fiber "github.com/gofiber/fiber/v2"
	qrcode "github.com/yeqown/go-qrcode/v2"
	standard "github.com/yeqown/go-qrcode/writer/standard"
)

// AdminUserList returns lightweight rows for every user.
func (s *Server) AdminUserList(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"result": s.Users.Summaries()})
}

// AdminUserSearchExact resolves a UUID from an exact display name.
func (s *Server) AdminUserSearchExact(c *fiber.Ctx) error {
	id := s.Users.SearchExact(c.Params("name"))
	if id == "" {
		return c.JSON(fiber.Map{"result": "not found"})
	}
	return c.JSON(fiber.Map{"result": id})
}

// AdminUserSearchFuzzy fuzzy-matches a name and returns full user records.
func (s *Server) AdminUserSearchFuzzy(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"result": s.Users.SearchFuzzy(c.Params("query"))})
}

// AdminUserByBarcode resolves a user from a (real or virtual) barcode.
func (s *Server) AdminUserByBarcode(c *fiber.Ctx) error {
	u, ok := s.Users.GetByBarcode(c.Params("barcode"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown barcode"})
	}
	return c.JSON(u)
}

// AdminUserCreate creates a user from a posted user JSON. A blank name becomes a
// numbered "Temp" user, matching v1's onboarding behavior.
func (s *Server) AdminUserCreate(c *fiber.Ctx) error {
	var in user.User
	json.Unmarshal(c.Body(), &in)

	if in.Identity.FirstName == "" && in.Identity.MiddleName == "" && in.Identity.LastName == "" {
		in.Identity.FirstName = "Temp"
		in.Identity.MiddleName = fmt.Sprintf("%06d", rand.Intn(1000000))
		in.Identity.LastName = fmt.Sprintf("%06d", rand.Intn(1000000))
	}

	created, err := s.Users.Create("temp")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	in.UUID = created.UUID
	in.CreatedDate = created.CreatedDate
	in.CreatedTime = created.CreatedTime
	if in.FamilySize < 1 {
		in.FamilySize = 1
	}
	user.ApplyBalance(&in, s.Cfg.Snapshot().Balance, in.FamilySize)
	if err := s.Users.Save(&in, user.SaveOptions{Remote: true}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	logger.GetLogger().Info(fmt.Sprintf("admin created user %s (%s)", in.NameString, in.UUID))
	return c.JSON(fiber.Map{"result": in})
}

// AdminUserEdit saves a full edited user record.
func (s *Server) AdminUserEdit(c *fiber.Ctx) error {
	u, err := s.Users.Update(c.Body())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	logger.GetLogger().Info(fmt.Sprintf("admin edited user %s", u.UUID))
	return c.JSON(fiber.Map{"result": true, "user": u})
}

// nopWriteCloser adapts a Writer to the io.WriteCloser the QR writer expects.
type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

// AdminHandoffQR renders the onboarding QR pointing at the user's fresh-login URL.
func (s *Server) AdminHandoffQR(c *fiber.Ctx) error {
	snap := s.Cfg.Snapshot()
	base := snap.ServerBaseURL
	if base == "" {
		base = snap.LocalHostURL
	}
	target := fmt.Sprintf("%s/user/login/fresh/%s", base, c.Params("uuid"))

	qrc, err := qrcode.New(target)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	var buf bytes.Buffer
	w := standard.NewWithWriter(nopWriteCloser{&buf})
	if err := qrc.Save(w); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", http.DetectContentType(buf.Bytes()))
	return c.Send(buf.Bytes())
}
