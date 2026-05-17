package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/models"
	"github.com/dadyutenga/hms-control/internal/views/client"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) ClientBillingPage(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}
	user, err := q.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(500).SendString("Failed to load user")
	}

	txns, err := q.GetBillingTransactionsByTenantID(c.Context(), tenant.ID.String())
	if err != nil {
		return err
	}

	instances, err := q.ListInstancesByTenantID(c.Context(), tenant.ID)
	if err != nil {
		instances = []generated.Instance{}
	}

	var unpaidInstances []generated.Instance
	for _, inst := range instances {
		if inst.BillingStatus == "unpaid" {
			unpaidInstances = append(unpaidInstances, inst)
		}
	}

	pmStore := models.NewPaymentMethodStore(h.db)
	paymentMethods, err := pmStore.ListActive()
	if err != nil {
		paymentMethods = []models.PaymentMethod{}
	}

	return render(c, client.BillingPage(client.BillingPageProps{
		Tenant:          tenant,
		User:            user,
		Transactions:    txns,
		UnpaidInstances: unpaidInstances,
		PaymentMethods:  paymentMethods,
	}))
}

func (h *Handler) ClientSubmitPayment(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	amountStr := c.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Status(400).SendString("Invalid amount")
	}

	method := c.FormValue("payment_method")
	reference := c.FormValue("reference")
	notes := c.FormValue("notes")

	description := fmt.Sprintf("Payment via %s", method)
	if reference != "" {
		description += " | Ref: " + reference
	}
	if notes != "" {
		description += " | " + notes
	}

	if _, err := q.CreateBillingTransaction(c.Context(), tenant.ID.String(), amount, generated.TxnTypePayment, description, nil); err != nil {
		return c.Status(500).SendString("Failed to submit payment")
	}

	return c.Redirect("/dashboard/billing")
}

func (h *Handler) ClientPayInstance(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	instID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	inst, err := q.GetInstanceByID(c.Context(), instID)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}
	if inst.TenantID != tenant.ID {
		return c.Status(403).SendString("Access denied")
	}

	pmID := c.FormValue("payment_method_id")
	reference := c.FormValue("reference")
	notes := c.FormValue("notes")

	// Get payment method name
	pmStore := models.NewPaymentMethodStore(h.db)
	pmName := "unknown"
	if pmID != "" {
		if id, err := strconv.ParseInt(pmID, 10, 64); err == nil {
			if pm, err := pmStore.GetByID(id); err == nil {
				pmName = pm.Name
			}
		}
	}

	description := fmt.Sprintf("Instance payment for %s via %s", inst.HotelName, pmName)
	if reference != "" {
		description += " | Ref: " + reference
	}
	if notes != "" {
		description += " | " + notes
	}

	// Handle receipt upload
	file, ferr := c.FormFile("receipt")
	if ferr == nil && file.Size <= 10<<20 {
		uploadDir := filepath.Join(h.cfg.UploadDir, "receipts", tenant.ID.String())
		if err := os.MkdirAll(uploadDir, 0755); err == nil {
			ext := filepath.Ext(file.Filename)
			filename := "receipt_" + randomHexName() + ext
			dst := filepath.Join(uploadDir, filename)
			if err := c.SaveFile(file, dst); err == nil {
				receiptPath := "uploads/receipts/" + tenant.ID.String() + "/" + filename
				description += " | File: " + receiptPath
			}
		}
	}

	// Create pending transaction (DO NOT mark as paid yet)
	q.CreateBillingTransaction(c.Context(), tenant.ID.String(), inst.Price, generated.TxnTypePayment, description, nil)

	return c.Redirect("/dashboard/instances/" + inst.ID.String())
}

func (h *Handler) ClientUploadReceipt(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	file, err := c.FormFile("receipt")
	if err != nil {
		return c.Status(400).SendString("No receipt file uploaded")
	}

	if file.Size > 10<<20 {
		return c.Status(400).SendString("File too large (max 10MB)")
	}

	uploadDir := filepath.Join(h.cfg.UploadDir, "receipts", tenant.ID.String())
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(500).SendString("Failed to save receipt")
	}

	ext := filepath.Ext(file.Filename)
	filename := "receipt_" + randomHexName() + ext
	dst := filepath.Join(uploadDir, filename)
	if err := c.SaveFile(file, dst); err != nil {
		return c.Status(500).SendString("Failed to save receipt")
	}

	reference := c.FormValue("reference")
	notes := c.FormValue("notes")
	description := "Receipt uploaded"
	if reference != "" {
		description += " | Ref: " + reference
	}
	if notes != "" {
		description += " | " + notes
	}

	receiptPath := "uploads/receipts/" + tenant.ID.String() + "/" + filename
	description += " | File: " + receiptPath

	if _, err := q.CreateBillingTransaction(c.Context(), tenant.ID.String(), 0, generated.TxnTypePayment, description, nil); err != nil {
		return c.Status(500).SendString("Failed to record receipt")
	}

	return c.Redirect("/dashboard/billing")
}

func randomHexName() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = "0123456789abcdef"[i%16]
	}
	return fmt.Sprintf("%x", b)
}
