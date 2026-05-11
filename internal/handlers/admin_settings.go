package handlers

import (
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
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

	user, ok := middleware.GetUser(c)
	if ok {
		h.audit.Log(user.ID, "settings.contact", nil, "Contact settings updated", c.IP())
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

	user, ok := middleware.GetUser(c)
	if ok {
		tid := id.String()
		h.audit.Log(user.ID, "billing.updated", &tid, status, c.IP())
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

	user, ok := middleware.GetUser(c)
	if ok {
		tid := id.String()
		h.audit.Log(user.ID, "deployment."+action, &tid, "", c.IP())
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

func (h *Handler) AdminSMTPSettings(c *fiber.Ctx) error {
	settings, err := h.settings.All()
	if err != nil {
		return err
	}
	saved := c.Query("saved") == "1"
	return render(c, admin.SMTPSettingsPage(settings, saved))
}

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
	for key, value := range fields {
		if err := h.settings.Set(key, value); err != nil {
			return err
		}
	}
	if user, ok := middleware.GetUser(c); ok {
		h.audit.Log(user.ID, "settings.smtp", nil, "SMTP settings updated", c.IP())
	}
	return c.Redirect("/admin/settings/smtp?saved=1")
}

func (h *Handler) TestSMTP(c *fiber.Ctx) error {
	to := c.FormValue("test_email")
	if to == "" {
		return c.JSON(fiber.Map{"ok": false, "error": "test_email is required"})
	}

	host := h.settings.Get("smtp_host")
	port := h.settings.Get("smtp_port")
	user := h.settings.Get("smtp_user")
	pass := h.settings.Get("smtp_pass")
	from := h.settings.Get("smtp_from")
	tlsEnabled := h.settings.Get("smtp_tls") != "false"

	if host == "" {
		return c.JSON(fiber.Map{"ok": false, "error": "SMTP host not configured"})
	}
	addr := net.JoinHostPort(host, port)
	var auth smtp.Auth
	if user != "" || pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: HMS SMTP Test\r\n\r\nSMTP is working correctly.",
		from, to,
	)
	var err error
	if tlsEnabled {
		err = smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	} else {
		var client *smtp.Client
		conn, dialErr := net.Dial("tcp", addr)
		if dialErr != nil {
			return c.JSON(fiber.Map{"ok": false, "error": dialErr.Error()})
		}
		client, err = smtp.NewClient(conn, host)
		if err == nil && auth != nil {
			if authErr := client.Auth(auth); authErr != nil {
				err = authErr
			}
		}
		if err == nil {
			if err = client.Mail(from); err == nil {
				if err = client.Rcpt(to); err == nil {
					var w io.WriteCloser
					w, err = client.Data()
					if err == nil {
						_, err = w.Write([]byte(msg))
						if closeErr := w.Close(); err == nil {
							err = closeErr
						}
					}
				}
			}
		}
		if client != nil {
			_ = client.Quit()
		}
	}
	if err != nil {
		return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) AdminProvisionerSettings(c *fiber.Ctx) error {
	settings, err := h.settings.All()
	if err != nil {
		return err
	}
	templates, err := h.templates.List()
	if err != nil {
		return err
	}
	saved := c.Query("saved") == "1"
	return render(c, admin.ProvisionerSettingsPage(settings, templates, saved, ""))
}

func (h *Handler) UpdateProvisionerSettings(c *fiber.Ctx) error {
	provisionScript := c.FormValue("provision_script")
	dockerTemplate := c.FormValue("docker_template")
	provisionTimeout := strings.TrimSpace(c.FormValue("provision_timeout"))

	if provisionTimeout != "" {
		if _, err := strconv.Atoi(provisionTimeout); err != nil {
			settings, _ := h.settings.All()
			templates, _ := h.templates.List()
			return render(c, admin.ProvisionerSettingsPage(settings, templates, false, "Provision timeout must be a number of seconds."))
		}
	}
	if dockerTemplate != "" {
		if _, err := h.templates.GetByName(dockerTemplate); err != nil {
			settings, _ := h.settings.All()
			templates, _ := h.templates.List()
			return render(c, admin.ProvisionerSettingsPage(settings, templates, false, "Selected docker template not found."))
		}
	}

	for key, value := range map[string]string{
		"provision_script":  provisionScript,
		"docker_template":   dockerTemplate,
		"provision_timeout": provisionTimeout,
	} {
		if err := h.settings.Set(key, value); err != nil {
			return err
		}
	}
	if user, ok := middleware.GetUser(c); ok {
		h.audit.Log(user.ID, "settings.provisioner", nil, "Provisioner settings updated", c.IP())
	}
	return c.Redirect("/admin/settings/provisioner?saved=1")
}
