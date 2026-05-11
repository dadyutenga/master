package handlers

import (
	"time"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) ImpersonateTenant(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	user, ok := middleware.GetUser(c)
	if !ok {
		return c.Redirect("/login")
	}

	sess.Set("original_admin_id", user.ID)
	sess.Set("impersonating_tenant_id", id.String())

	tid := id.String()
	LogAction(h.db, user.ID, "tenant.impersonated", &tid, "", c.IP())

	sess.Set("userID", tenant.UserID)
	if err := sess.Save(); err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:    "hms_impersonated",
		Value:   id.String(),
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})

	return c.Redirect("/dashboard")
}

func (h *Handler) StopImpersonation(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}

	sess.Delete("impersonating_tenant_id")
	if origID := sess.Get("original_admin_id"); origID != nil {
		sess.Set("userID", origID)
		sess.Delete("original_admin_id")
	}
	if err := sess.Save(); err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:    "hms_impersonated",
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-1 * time.Hour),
	})

	return c.Redirect("/admin/tenants")
}
