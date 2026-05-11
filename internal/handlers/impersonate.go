package handlers

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// POST /admin/tenants/:id/impersonate
func (h *Handler) ImpersonateTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	var email string
	var userID int64
	err = h.db.QueryRow(
		`SELECT u.id, u.email FROM users u
		 JOIN tenants t ON t.user_id = u.id
		 WHERE t.id = ? AND u.role = 'client'`, id.String(),
	).Scan(&userID, &email)
	if err == sql.ErrNoRows {
		return c.Status(404).SendString("tenant user not found")
	}
	if err != nil {
		return fmt.Errorf("impersonate lookup: %w", err)
	}

	sess, err := h.store.Get(c)
	if err != nil {
		return err
	}

	adminID := c.Locals("user_id").(int64)

	sess.Set("original_admin_id", adminID)
	sess.Set("impersonating_tenant_id", id.String())
	sess.Set("userID", userID)
	if err := sess.Save(); err != nil {
		return err
	}

	tid := id.String()
	h.audit.Log(adminID, "tenant.impersonated", &tid,
		fmt.Sprintf("impersonating %s", email), c.IP())

	c.Cookie(&fiber.Cookie{
		Name:    "hms_impersonated",
		Value:   id.String(),
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})

	return c.Redirect("/dashboard")
}

// POST /impersonate/stop
func (h *Handler) StopImpersonation(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}

	origID := sess.Get("original_admin_id")
	if origID != nil {
		sess.Set("userID", origID)
	}
	sess.Delete("impersonating_tenant_id")
	sess.Delete("original_admin_id")
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
