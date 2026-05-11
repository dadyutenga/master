package handlers

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/models"
	"github.com/dadyutenga/hms-control/internal/views/admin"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
	var stats models.DashboardStats

	err := h.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status='active'    THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='pending_verification' OR status='pending_approval' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='suspended' THEN 1 ELSE 0 END), 0)
		FROM tenants
	`).Scan(
		&stats.TotalTenants,
		&stats.ActiveTenants,
		&stats.PendingTenants,
		&stats.SuspendedTenants,
	)
	if err != nil {
		return err
	}

	err = h.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN status='active' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='failed' THEN 1 ELSE 0 END), 0)
		FROM deployments
	`).Scan(
		&stats.RunningDeployments,
		&stats.FailedDeployments,
	)
	if err != nil {
		return err
	}

	recentPage, _ := h.audit.List(1, 10, "", "")
	stats.RecentActions = recentPage.Logs

	return render(c, admin.DashboardPage(stats))
}

func (h *Handler) ListTenants(c *fiber.Ctx) error {
	status := c.Query("status", "")
	search := c.Query("q", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	tenants, totalPages, err := h.listTenantsWithFilters(status, search, page)
	if err != nil {
		return err
	}
	return render(c, admin.TenantList(tenants, search, status, page, totalPages))
}

func (h *Handler) ListVerificationTenants(c *fiber.Ctx) error {
	status := c.Query("status", "pending_verification,pending_approval")
	search := c.Query("q", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	tenants, totalPages, err := h.listTenantsWithFilters(status, search, page)
	if err != nil {
		return err
	}
	return render(c, admin.VerificationList(tenants, search, status, page, totalPages))
}

func (h *Handler) ShowVerificationTenant(c *fiber.Ctx) error {
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

	documents, err := q.GetDocumentsByUserID(c.Context(), tenant.UserID)
	if err != nil {
		return err
	}

	return render(c, admin.VerificationDetail(tenant, user, documents))
}

func (h *Handler) VerifyTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}
	if tenant.Status != generated.TenantStatusPendingVerification {
		return c.Status(409).SendString("Tenant not in pending_verification state")
	}

	if err := q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
		ID:     id,
		Status: generated.TenantStatusPendingApproval,
	}); err != nil {
		return c.Status(500).SendString("Failed to verify tenant.")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "tenant.verified", &tid, "", c.IP())
	}

	if c.Get("HX-Request") == "true" {
		updatedTenant, _ := q.GetTenantByID(c.Context(), id)
		user, _ := q.GetUserByID(c.Context(), updatedTenant.UserID)
		userName, userEmail := "", ""
		if user.Name != "" {
			userName = user.Name
		}
		if user.Email != "" {
			userEmail = user.Email
		}
		row := generated.ListTenantsRow{
			ID:            updatedTenant.ID,
			UserID:        updatedTenant.UserID,
			CompanyName:   updatedTenant.CompanyName,
			Slug:          updatedTenant.Slug,
			Domain:        updatedTenant.Domain,
			DbName:        updatedTenant.DbName,
			DbUser:        updatedTenant.DbUser,
			DbPassword:    updatedTenant.DbPassword,
			Status:        updatedTenant.Status,
			BillingStatus: updatedTenant.BillingStatus,
			CreatedAt:     updatedTenant.CreatedAt,
			UpdatedAt:     updatedTenant.UpdatedAt,
			UserName:      userName,
			UserEmail:     userEmail,
		}
		return render(c, admin.VerificationRow(row))
	}

	redirectTo := c.Get("Referer")
	if redirectTo == "" {
		redirectTo = "/admin/verification"
	}
	return c.Redirect(redirectTo)
}

func (h *Handler) listTenantsWithFilters(status, search string, page int) ([]generated.ListTenantsRow, int, error) {
	const limit = 20
	offset := (page - 1) * limit

	baseWhere := "WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		// Support multiple statuses separated by comma (e.g., "pending_verification,pending_approval")
		statuses := strings.Split(status, ",")
		placeholders := make([]string, len(statuses))
		for i, s := range statuses {
			placeholders[i] = "?"
			args = append(args, strings.TrimSpace(s))
		}
		baseWhere += " AND t.status IN (" + strings.Join(placeholders, ",") + ")"
	}
	if search != "" {
		baseWhere += " AND (t.company_name LIKE ? OR u.email LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	countQuery := `SELECT COUNT(*) FROM tenants t JOIN users u ON t.user_id = u.id ` + baseWhere
	var total int
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT t.id, t.user_id, t.company_name, t.slug, t.domain, t.db_name, t.db_user, t.db_password, t.app_key, t.status, t.provision_log, t.approved_at, t.provisioned_at, t.billing_status, t.created_at, t.updated_at, u.name as user_name, u.email as user_email
		FROM tenants t
		JOIN users u ON t.user_id = u.id ` + baseWhere + `
		ORDER BY
		    CASE t.status
		        WHEN 'pending_approval' THEN 1
		        WHEN 'pending_verification' THEN 2
		        WHEN 'provisioning'     THEN 3
		        WHEN 'failed'           THEN 4
		        WHEN 'active'           THEN 5
		        ELSE 6
		    END,
		    t.created_at DESC
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tenants []generated.ListTenantsRow
	for rows.Next() {
		var r generated.ListTenantsRow
		var appKey, provisionLog sql.NullString
		var approvedAt, provisionedAt sql.NullTime
		err := rows.Scan(
			&r.ID, &r.UserID, &r.CompanyName, &r.Slug, &r.Domain,
			&r.DbName, &r.DbUser, &r.DbPassword, &appKey, &r.Status,
			&provisionLog, &approvedAt, &provisionedAt, &r.BillingStatus, &r.CreatedAt, &r.UpdatedAt,
			&r.UserName, &r.UserEmail,
		)
		if err != nil {
			return nil, 0, err
		}
		if appKey.Valid {
			r.AppKey = &appKey.String
		}
		if provisionLog.Valid {
			r.ProvisionLog = &provisionLog.String
		}
		if approvedAt.Valid {
			r.ApprovedAt = &approvedAt.Time
		}
		if provisionedAt.Valid {
			r.ProvisionedAt = &provisionedAt.Time
		}
		tenants = append(tenants, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	return tenants, totalPages, nil
}

func (h *Handler) ShowTenant(c *fiber.Ctx) error {
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

	deployments, err := q.ListDeploymentsByTenantID(c.Context(), tenant.ID)
	if err != nil {
		return err
	}

	documents, err := q.GetDocumentsByUserID(c.Context(), tenant.UserID)
	if err != nil {
		return err
	}

	return render(c, admin.TenantDetail(tenant, user, deployments, documents))
}

func (h *Handler) ApproveTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}
	if tenant.Status != generated.TenantStatusPendingApproval {
		return c.Status(409).SendString("Tenant not in pending_approval state")
	}

	if err := q.ApproveTenant(c.Context(), id); err != nil {
		return c.Status(500).SendString("Failed to approve tenant.")
	}

	if _, err := q.CreateDeployment(c.Context(), generated.CreateDeploymentParams{
		TenantID: id,
		Action:   "approve",
		Status:   generated.DeploymentStatusProvisioning,
	}); err != nil {
		return c.Status(500).SendString("Failed to record deployment.")
	}

	h.eng.Enqueue(id)

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "tenant.approved", &tid, "", c.IP())
	}

	if c.Get("HX-Request") == "true" {
		referer := c.Get("Referer")
		if strings.Contains(referer, "/admin/verification") {
			updatedTenant, _ := q.GetTenantByID(c.Context(), id)
			user, _ := q.GetUserByID(c.Context(), updatedTenant.UserID)
			userName, userEmail := "", ""
			if user.Name != "" {
				userName = user.Name
			}
			if user.Email != "" {
				userEmail = user.Email
			}
			row := generated.ListTenantsRow{
				ID:            updatedTenant.ID,
				UserID:        updatedTenant.UserID,
				CompanyName:   updatedTenant.CompanyName,
				Slug:          updatedTenant.Slug,
				Domain:        updatedTenant.Domain,
				DbName:        updatedTenant.DbName,
				DbUser:        updatedTenant.DbUser,
				DbPassword:    updatedTenant.DbPassword,
				Status:        updatedTenant.Status,
				BillingStatus: updatedTenant.BillingStatus,
				CreatedAt:     updatedTenant.CreatedAt,
				UpdatedAt:     updatedTenant.UpdatedAt,
				UserName:      userName,
				UserEmail:     userEmail,
			}
			return render(c, admin.VerificationRow(row))
		}
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}

	redirectTo := c.Get("Referer")
	if redirectTo == "" {
		redirectTo = "/admin/tenants"
	}
	return c.Redirect(redirectTo)
}

func (h *Handler) SuspendTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}
	if tenant.Status != generated.TenantStatusActive {
		return c.Status(409).SendString("Only active tenants can be suspended")
	}

	if err := q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
		ID:     id,
		Status: generated.TenantStatusSuspended,
	}); err != nil {
		return c.Status(500).SendString("Failed to suspend tenant.")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "tenant.suspended", &tid, "", c.IP())
	}

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusSuspended))
	}
	return c.Redirect("/admin/tenants")
}

func (h *Handler) RetryProvision(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}
	if tenant.Status != generated.TenantStatusFailed {
		return c.Status(409).SendString("Only failed tenants can be retried")
	}

	if err := q.SetTenantProvisioning(c.Context(), id); err != nil {
		return c.Status(500).SendString("Failed to reset tenant status.")
	}
	if _, err := q.CreateDeployment(c.Context(), generated.CreateDeploymentParams{
		TenantID: id,
		Action:   "retry",
		Status:   generated.DeploymentStatusProvisioning,
	}); err != nil {
		return c.Status(500).SendString("Failed to record deployment.")
	}
	h.eng.Enqueue(id)

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "provision.retry", &tid, "", c.IP())
	}

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}
	return c.Redirect("/admin/tenants/" + id.String())
}

// GET /admin/audit
func (h *Handler) AuditLog(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	action := c.Query("action", "")
	search := c.Query("q", "")
	if page < 1 {
		page = 1
	}

	result, err := h.audit.List(page, 50, action, search)
	if err != nil {
		return fmt.Errorf("audit list: %w", err)
	}
	return render(c, admin.AuditLogPage(result, action, search))
}

// GET /admin/audit/export
func (h *Handler) ExportAuditCSV(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="audit-%s.csv"`, time.Now().Format("2006-01-02")))

	result, err := h.audit.List(1, 100000, "", "")
	if err != nil {
		return err
	}

	w := csv.NewWriter(c.Response().BodyWriter())
	w.Write([]string{"ID", "Admin", "Action", "Tenant ID", "Tenant", "Detail", "IP", "Created At"})
	for _, l := range result.Logs {
		tid, tname := "", ""
		if l.TenantID != nil {
			tid = *l.TenantID
		}
		if l.TenantName != nil {
			tname = *l.TenantName
		}
		w.Write([]string{
			strconv.FormatInt(l.ID, 10),
			l.AdminEmail, l.Action, tid, tname,
			l.Detail, l.IPAddress, l.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()
	return w.Error()
}

func (h *Handler) TenantHealthCheck(c *fiber.Ctx) error {
	tenantID := c.Params("id")

	var endpoint string
	err := h.db.QueryRow(
		`SELECT COALESCE(d.endpoint, '') FROM deployments d
		 WHERE d.tenant_id = ? AND d.status = 'active'
		 ORDER BY d.created_at DESC LIMIT 1`,
		tenantID,
	).Scan(&endpoint)

	if err == sql.ErrNoRows || endpoint == "" {
		var domain string
		h.db.QueryRow(`SELECT domain FROM tenants WHERE id = ?`, tenantID).Scan(&domain)
		if domain == "" {
			return c.JSON(fiber.Map{"status": "UNKNOWN", "tenant_id": tenantID, "error": "no deployment or domain found"})
		}
		endpoint = "https://" + domain
	}

	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	resp, reqErr := client.Get(endpoint + "/health")
	latency := time.Since(start).Milliseconds()

	if reqErr != nil {
		return c.JSON(fiber.Map{
			"status": "DOWN", "tenant_id": tenantID,
			"endpoint": endpoint, "latency_ms": latency,
			"error": reqErr.Error(),
		})
	}
	resp.Body.Close()

	status := "DOWN"
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = "UP"
	}
	return c.JSON(fiber.Map{
		"status": status, "tenant_id": tenantID,
		"endpoint": endpoint, "status_code": resp.StatusCode, "latency_ms": latency,
	})
}

func (h *Handler) StreamProvisionLogs(c *fiber.Ctx) error {
	tenantID := c.Params("id")
	logPath := fmt.Sprintf("./tmp/provision-%s.log", tenantID)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		f, err := os.Open(logPath)
		if err != nil {
			fmt.Fprintf(w, "data: [error] log file not found for tenant %s\n\n", tenantID)
			w.Flush()
			return
		}
		defer f.Close()

		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()

		reader := bufio.NewReaderSize(f, 4096)
		for {
			select {
			case <-heartbeat.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				w.Flush()
			default:
				line, err := reader.ReadString('\n')
				if len(line) > 0 {
					line = strings.TrimRight(line, "\r\n")
					fmt.Fprintf(w, "data: %s\n\n", line)
					w.Flush()
				}
				if err != nil {
					if err.Error() == "EOF" {
						var provStatus string
						h.db.QueryRow(
							`SELECT status FROM deployments WHERE tenant_id = ?
							 ORDER BY created_at DESC LIMIT 1`, tenantID,
						).Scan(&provStatus)

						if provStatus == "active" || provStatus == "failed" {
							fmt.Fprintf(w, "event: done\ndata: %s\n\n", provStatus)
							w.Flush()
							return
						}
						time.Sleep(500 * time.Millisecond)
						continue
					}
					fmt.Fprintf(w, "data: [stream error] %s\n\n", err.Error())
					w.Flush()
					return
				}
			}
		}
	})

	return nil
}

func (h *Handler) ExportTenantsCSV(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="tenants-%s.csv"`, time.Now().Format("2006-01-02")))

	rows, err := h.db.Query(`
		SELECT
			t.id,
			t.company_name,
			u.email,
			t.status,
			COALESCE(t.billing_status, 'none'),
			COALESCE(t.domain, ''),
			t.created_at
		FROM tenants t
		JOIN users u ON t.user_id = u.id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := csv.NewWriter(c.Response().BodyWriter())
	w.Write([]string{"ID", "Name", "Email", "Status", "Billing", "Domain", "Registered At"})

	for rows.Next() {
		var id, name, email, status, billing, domain, createdAt string
		rows.Scan(&id, &name, &email, &status, &billing, &domain, &createdAt)
		w.Write([]string{id, name, email, status, billing, domain, createdAt})
	}

	w.Flush()
	return w.Error()
}
