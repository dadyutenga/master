package handlers

import (
	"strconv"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/models"
	"github.com/dadyutenga/hms-control/internal/views/admin"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ── Payment Methods Settings ─────────────────────────────────────────

func (h *Handler) AdminPaymentSettings(c *fiber.Ctx) error {
	store := models.NewPaymentMethodStore(h.db)
	methods, err := store.List()
	if err != nil {
		methods = []models.PaymentMethod{}
	}
	saved := c.Query("saved") == "1"
	return render(c, admin.PaymentSettingsPage(admin.PaymentSettingsProps{
		Methods: methods,
		Saved:   saved,
	}))
}

func (h *Handler) AdminCreatePaymentMethod(c *fiber.Ctx) error {
	name := c.FormValue("name")
	methodType := c.FormValue("method_type")
	apiKey := c.FormValue("api_key")
	apiSecret := c.FormValue("api_secret")
	webhookSecret := c.FormValue("webhook_secret")
	callbackURL := c.FormValue("callback_url")
	lipaNamba := c.FormValue("lipa_namba")

	if name == "" || methodType == "" {
		return c.Status(400).SendString("Name and type are required")
	}

	store := models.NewPaymentMethodStore(h.db)
	if _, err := store.Create(name, methodType, apiKey, apiSecret, webhookSecret, callbackURL, lipaNamba); err != nil {
		return c.Status(500).SendString("Failed to create payment method: " + err.Error())
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "payment_method.created", nil, "created method '"+name+"'", c.IP())
	}

	return c.Redirect("/admin/settings/payment-methods?saved=1")
}

func (h *Handler) AdminEditPaymentMethod(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	store := models.NewPaymentMethodStore(h.db)
	m, err := store.GetByID(id)
	if err != nil {
		return fiber.ErrNotFound
	}

	return render(c, admin.PaymentMethodEditPage(*m))
}

func (h *Handler) AdminUpdatePaymentMethod(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	name := c.FormValue("name")
	methodType := c.FormValue("method_type")
	apiKey := c.FormValue("api_key")
	apiSecret := c.FormValue("api_secret")
	webhookSecret := c.FormValue("webhook_secret")
	callbackURL := c.FormValue("callback_url")
	lipaNamba := c.FormValue("lipa_namba")

	if name == "" || methodType == "" {
		return c.Status(400).SendString("Name and type are required")
	}

	store := models.NewPaymentMethodStore(h.db)
	if err := store.Update(id, name, methodType, apiKey, apiSecret, webhookSecret, callbackURL, lipaNamba); err != nil {
		return c.Status(500).SendString("Failed to update payment method")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "payment_method.updated", nil, "updated method '"+name+"'", c.IP())
	}

	return c.Redirect("/admin/settings/payment-methods?saved=1")
}

func (h *Handler) AdminTogglePaymentMethod(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	store := models.NewPaymentMethodStore(h.db)
	m, err := store.GetByID(id)
	if err != nil {
		return fiber.ErrNotFound
	}

	if err := store.ToggleActive(id, !m.IsActive); err != nil {
		return c.Status(500).SendString("Failed to toggle payment method")
	}

	return c.Redirect("/admin/settings/payment-methods")
}

// ── Pending Payments Verification ─────────────────────────────────────

func (h *Handler) AdminPendingPayments(c *fiber.Ctx) error {
	q := generated.New(h.db)
	payments, err := q.ListAllPayments(c.Context())
	if err != nil {
		return c.Status(500).SendString("Failed to load payments: " + err.Error())
	}

	var pending []generated.ListPaymentsRow
	for _, p := range payments {
		if p.Status == "pending" {
			pending = append(pending, p)
		}
	}
	if pending == nil {
		pending = []generated.ListPaymentsRow{}
	}

	return render(c, admin.PendingPaymentsPage(admin.PendingPaymentsProps{
		Payments: pending,
	}))
}

func (h *Handler) AdminApprovePayment(c *fiber.Ctx) error {
	idStr := c.Params("id")
	txnID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid transaction ID")
	}

	q := generated.New(h.db)
	txn, err := q.GetBillingTransactionByID(c.Context(), txnID)
	if err != nil {
		return c.Status(404).SendString("Transaction not found")
	}

	// Mark transaction as completed
	q.CreateBillingTransaction(c.Context(), txn.TenantID, 0, generated.TxnTypeAdjustment,
		"Payment approved by admin", nil)

	// Mark all unpaid instances for this tenant as paid and start provisioning
	tenantID, err := uuid.Parse(txn.TenantID)
	if err == nil {
		instances, _ := q.ListInstancesByTenantID(c.Context(), tenantID)
		for _, inst := range instances {
			if inst.PaymentStatus == "pending" || inst.BillingStatus == "unpaid" || inst.BillingStatus == "overdue" {
				q.UpdateInstancePaymentStatus(c.Context(), inst.ID, "")
				q.MarkInstancePaid(c.Context(), inst.ID)
				q.UpdateInstanceStatus(c.Context(), inst.ID, "provisioning")
				q.CreateInstanceDeployment(c.Context(), inst.ID, "provision", "provisioning")
			}
		}

		q.MarkTenantPaid(c.Context(), txn.TenantID)
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := txn.TenantID
		h.audit.Log(uid, "payment.approved", &tid, "approved payment transaction "+idStr, c.IP())
	}

	return c.Redirect("/admin/payments/pending")
}

func (h *Handler) AdminRejectPayment(c *fiber.Ctx) error {
	idStr := c.Params("id")
	txnID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid transaction ID")
	}

	q := generated.New(h.db)
	txn, err := q.GetBillingTransactionByID(c.Context(), txnID)
	if err != nil {
		return c.Status(404).SendString("Transaction not found")
	}

	q.CreateBillingTransaction(c.Context(), txn.TenantID, 0, generated.TxnTypeAdjustment,
		"Payment rejected by admin", nil)

	// Clear payment_status on instances for this tenant
	tenantID, err := uuid.Parse(txn.TenantID)
	if err == nil {
		instances, _ := q.ListInstancesByTenantID(c.Context(), tenantID)
		for _, inst := range instances {
			if inst.PaymentStatus == "pending" {
				q.UpdateInstancePaymentStatus(c.Context(), inst.ID, "")
			}
		}
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := txn.TenantID
		h.audit.Log(uid, "payment.rejected", &tid, "rejected payment transaction "+idStr, c.IP())
	}

	return c.Redirect("/admin/payments/pending")
}
