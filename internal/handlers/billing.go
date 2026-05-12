package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/views/admin"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) TenantBillingPage(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	user, err := q.GetUserByID(c.Context(), tenant.UserID)
	if err != nil {
		return fiber.ErrNotFound
	}

	txns, err := q.GetBillingTransactionsByTenantID(c.Context(), id.String())
	if err != nil {
		return err
	}

	return render(c, admin.BillingPage(tenant, user, txns))
}

func (h *Handler) ChargeTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	amountStr := c.FormValue("amount")
	description := c.FormValue("description")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Status(400).SendString("Invalid amount")
	}
	if description == "" {
		description = "Service charge"
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	var adminID *int64
	if uid, ok := middleware.GetUserID(c); ok {
		adminID = &uid
	}

	if _, err := q.CreateBillingTransaction(c.Context(), tenant.ID.String(), amount, generated.TxnTypeCharge, description, adminID); err != nil {
		return c.Status(500).SendString("Failed to create charge")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		detail := fmt.Sprintf("charged %.0f TZS - %s", amount, description)
		h.audit.Log(uid, "billing.charge", &tid, detail, c.IP())
	}

	return c.Redirect("/admin/tenants/" + id.String() + "/billing")
}

func (h *Handler) RecordPayment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	amountStr := c.FormValue("amount")
	description := c.FormValue("description")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Status(400).SendString("Invalid amount")
	}
	if description == "" {
		description = "Payment received"
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	var adminID *int64
	if uid, ok := middleware.GetUserID(c); ok {
		adminID = &uid
	}

	if _, err := q.CreateBillingTransaction(c.Context(), tenant.ID.String(), amount, generated.TxnTypePayment, description, adminID); err != nil {
		return c.Status(500).SendString("Failed to record payment")
	}

	if err := q.MarkTenantPaid(c.Context(), tenant.ID.String()); err != nil {
		return c.Status(500).SendString("Failed to update billing status")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		detail := fmt.Sprintf("payment recorded %.0f TZS - %s", amount, description)
		h.audit.Log(uid, "billing.payment", &tid, detail, c.IP())
	}

	return c.Redirect("/admin/tenants/" + id.String() + "/billing")
}

func (h *Handler) UpdateBillingStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	status := c.FormValue("billing_status")
	if status != "unpaid" && status != "paid" && status != "overdue" && status != "suspended" {
		return c.Status(400).SendString("Invalid billing status")
	}

	q := generated.New(h.db)

	var lastPayment, nextDue *time.Time
	if lp := c.FormValue("last_payment_at"); lp != "" {
		t, err := time.Parse("2006-01-02", lp)
		if err == nil {
			lastPayment = &t
		}
	}
	if nd := c.FormValue("next_due_at"); nd != "" {
		t, err := time.Parse("2006-01-02", nd)
		if err == nil {
			nextDue = &t
		}
	}

	if err := q.UpdateTenantBilling(c.Context(), generated.UpdateTenantBillingParams{
		BillingStatus: status,
		LastPaymentAt: lastPayment,
		NextDueAt:     nextDue,
		ID:            id,
	}); err != nil {
		return c.Status(500).SendString("Failed to update billing")
	}

	return c.Redirect("/admin/tenants/" + id.String() + "/billing")
}
