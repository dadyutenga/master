package handlers

import (
	"database/sql"

	"github.com/dadyutenga/hms-control/internal/config"
	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/mailer"
	"github.com/dadyutenga/hms-control/internal/provisioner"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type Handler struct {
	cfg   *config.Config
	db    *sql.DB
	mail  *mailer.Mailer
	store *session.Store
	eng   *provisioner.Engine
}

func New(cfg *config.Config, db *sql.DB, mail *mailer.Mailer, store *session.Store, eng *provisioner.Engine) *Handler {
	return &Handler{
		cfg:   cfg,
		db:    db,
		mail:  mail,
		store: store,
		eng:   eng,
	}
}

func render(c *fiber.Ctx, comp templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return comp.Render(c.Context(), c.Response().BodyWriter())
}

func (h *Handler) contactDetails(c *fiber.Ctx) (generated.ContactDetails, error) {
	q := generated.New(h.db)
	return q.GetContactDetails(c.Context())
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.Status(code).SendString(err.Error())
}
