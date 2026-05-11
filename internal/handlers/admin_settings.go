package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/models"
	"github.com/dadyutenga/hms-control/internal/provisioner"
	"github.com/dadyutenga/hms-control/internal/views/admin"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) AdminContactSettings(c *fiber.Ctx) error {
	contact, err := h.contactDetails(c)
	if err != nil {
		return err
	}
	return render(c, admin.ContactSettings(admin.ContactSettingsProps{Contact: contact}))
}

func (h *Handler) UpdateContactSettings(c *fiber.Ctx) error {
	location := strings.TrimSpace(c.FormValue("location"))
	phone := strings.TrimSpace(c.FormValue("phone_number"))
	if location == "" || phone == "" {
		contact, _ := h.contactDetails(c)
		return render(c, admin.ContactSettings(admin.ContactSettingsProps{
			Contact: contact,
			Error:   "Location and phone number are required.",
		}))
	}

	q := generated.New(h.db)
	contact, err := q.UpsertContactDetails(c.Context(), generated.CreateContactDetailsParams{
		Location:    location,
		PhoneNumber: phone,
	})
	if err != nil {
		return render(c, admin.ContactSettings(admin.ContactSettingsProps{
			Contact: contact,
			Error:   "Failed to update contact details.",
		}))
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "settings.contact", nil, "", c.IP())
	}

	return render(c, admin.ContactSettings(admin.ContactSettingsProps{
		Contact: contact,
		Success: true,
	}))
}

func (h *Handler) UpdateTenantBilling(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	status := strings.TrimSpace(c.FormValue("billing_status"))
	if status != generated.BillingStatusPaid && status != generated.BillingStatusOverdue && status != generated.BillingStatusSuspended {
		return c.Status(400).SendString("Invalid billing status")
	}

	var lastPayment *time.Time
	lastPaymentRaw := strings.TrimSpace(c.FormValue("last_payment_at"))
	if lastPaymentRaw != "" {
		parsed, err := time.Parse("2006-01-02", lastPaymentRaw)
		if err != nil {
			return c.Status(400).SendString("Invalid last payment date")
		}
		lastPayment = &parsed
	}

	var nextDue *time.Time
	nextDueRaw := strings.TrimSpace(c.FormValue("next_due_at"))
	if nextDueRaw != "" {
		parsed, err := time.Parse("2006-01-02", nextDueRaw)
		if err != nil {
			return c.Status(400).SendString("Invalid next due date")
		}
		nextDue = &parsed
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	if err := q.UpdateTenantBilling(c.Context(), generated.UpdateTenantBillingParams{
		ID:            id,
		BillingStatus: status,
		LastPaymentAt: lastPayment,
		NextDueAt:     nextDue,
	}); err != nil {
		return c.Status(500).SendString("Failed to update billing status.")
	}

	if status == generated.BillingStatusOverdue || status == generated.BillingStatusSuspended {
		h.runDeploymentAction(c, tenant, "stop")
	}
	if status == generated.BillingStatusPaid {
		h.runDeploymentAction(c, tenant, "start")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "billing.updated", &tid, status, c.IP())
	}

	return c.Redirect("/admin/tenants/" + id.String())
}

func (h *Handler) StartTenantDeployment(c *fiber.Ctx) error {
	return h.handleDeploymentAction(c, "start")
}

func (h *Handler) StopTenantDeployment(c *fiber.Ctx) error {
	return h.handleDeploymentAction(c, "stop")
}

func (h *Handler) ShowDeployment(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}
	deploymentID, err := uuid.Parse(c.Params("deploymentId"))
	if err != nil {
		return c.Status(400).SendString("Invalid deployment ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), tenantID)
	if err != nil {
		return fiber.ErrNotFound
	}
	deployment, err := q.GetDeploymentByID(c.Context(), deploymentID)
	if err != nil || deployment.TenantID != tenantID {
		return fiber.ErrNotFound
	}

	return render(c, admin.DeploymentDetail(admin.DeploymentDetailProps{
		Tenant:     tenant,
		Deployment: deployment,
	}))
}

func (h *Handler) handleDeploymentAction(c *fiber.Ctx, action string) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	if _, err := h.runDeploymentAction(c, tenant, action); err != nil {
		return c.Status(500).SendString("Deployment action failed.")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "deployment."+action, &tid, "", c.IP())
	}

	return c.Redirect("/admin/tenants/" + id.String())
}

func (h *Handler) runDeploymentAction(c *fiber.Ctx, tenant generated.Tenant, action string) (*generated.Deployment, error) {
	q := generated.New(h.db)
	runner := provisioner.NewRunner(h.cfg)
	now := time.Now()

	var status string
	var logOutput string
	var errMsg string
	var err error

	switch action {
	case "start":
		status = generated.DeploymentStatusActive
		logOutput, err = runner.StartTenant(tenant.Slug)
	case "stop":
		status = generated.DeploymentStatusStopped
		logOutput, err = runner.StopTenant(tenant.Slug)
	default:
		return nil, fiber.ErrBadRequest
	}

	if err != nil {
		errMsg = err.Error()
		status = generated.DeploymentStatusFailed
	}

	deployment, depErr := q.CreateDeployment(c.Context(), generated.CreateDeploymentParams{
		TenantID:     tenant.ID,
		Action:       action,
		Status:       status,
		Log:          nullString(logOutput),
		ErrorMessage: nullString(errMsg),
		CompletedAt:  &now,
	})
	if depErr != nil {
		return nil, depErr
	}
	return &deployment, err
}

// ── SMTP Settings ──────────────────────────────────────────────────

// GET /admin/settings/smtp
func (h *Handler) AdminSMTPSettings(c *fiber.Ctx) error {
	s, err := h.settings.All()
	if err != nil {
		return err
	}
	saved := c.Query("saved") == "1"
	return render(c, admin.SMTPSettingsPage(s, saved))
}

// POST /admin/settings/smtp
func (h *Handler) UpdateSMTPSettings(c *fiber.Ctx) error {
	fields := map[string]string{
		"smtp_host": c.FormValue("smtp_host"),
		"smtp_port": c.FormValue("smtp_port"),
		"smtp_user": c.FormValue("smtp_user"),
		"smtp_from": c.FormValue("smtp_from"),
		"smtp_tls":  c.FormValue("smtp_tls"),
	}
	if pass := c.FormValue("smtp_pass"); pass != "" {
		fields["smtp_pass"] = pass
	}
	for k, v := range fields {
		if err := h.settings.Set(k, v); err != nil {
			return fmt.Errorf("save setting %s: %w", k, err)
		}
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "settings.smtp", nil, "SMTP settings updated", c.IP())
	}
	return c.Redirect("/admin/settings/smtp?saved=1")
}

// POST /admin/settings/smtp/test
func (h *Handler) TestSMTP(c *fiber.Ctx) error {
	to := c.FormValue("test_email")
	if to == "" {
		return c.JSON(fiber.Map{"ok": false, "error": "provide a test email address"})
	}
	err := h.mail.Send(to, "HMS SMTP Test", "If you see this, SMTP is configured correctly.")
	if err != nil {
		return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true, "message": "Test email sent to " + to})
}

// ── Provisioner Settings ───────────────────────────────────────────

// GET /admin/settings/provisioner
func (h *Handler) AdminProvisionerSettings(c *fiber.Ctx) error {
	s, err := h.settings.All()
	if err != nil {
		return err
	}
	saved := c.Query("saved") == "1"

	templateStore := models.NewDockerTemplateStore(h.db)
	templateNames, _ := templateStore.ListNames()

	return render(c, admin.ProvisionerSettingsPage(s, saved, templateNames))
}

// POST /admin/settings/provisioner
func (h *Handler) UpdateProvisionerSettings(c *fiber.Ctx) error {
	for _, k := range []string{"provision_script", "docker_template", "provision_timeout"} {
		if err := h.settings.Set(k, c.FormValue(k)); err != nil {
			return fmt.Errorf("save setting %s: %w", k, err)
		}
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "settings.provisioner", nil, "Provisioner settings updated", c.IP())
	}
	return c.Redirect("/admin/settings/provisioner?saved=1")
}
