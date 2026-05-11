package handlers

import (
	"database/sql"
	"encoding/csv"
	"strconv"

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

	rows, err := h.db.Query(`
		SELECT a.id, u.email, a.action, a.tenant_id, a.detail, a.ip_address, a.created_at
		FROM audit_logs a
		JOIN users u ON u.id = a.admin_id
		ORDER BY a.created_at DESC
		LIMIT 10
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(
			&l.ID, &l.AdminEmail, &l.Action,
			&l.TenantID, &l.Detail, &l.IPAddress, &l.CreatedAt,
		); err != nil {
			return err
		}
		stats.RecentActions = append(stats.RecentActions, l)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return render(c, admin.DashboardPage(stats))
}

func (h *Handler) ListTenants(c *fiber.Ctx) error {
	status := c.Query("status", "")
	search := c.Query("q", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	const limit = 20
	offset := (page - 1) * limit

	baseWhere := "WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		baseWhere += " AND t.status = ?"
		args = append(args, status)
	}
	if search != "" {
		baseWhere += " AND (t.company_name LIKE ? OR u.email LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	countQuery := `SELECT COUNT(*) FROM tenants t JOIN users u ON t.user_id = u.id ` + baseWhere
	var total int
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return err
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
		return err
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
			return err
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
		return err
	}

	totalPages := (total + limit - 1) / limit
	return render(c, admin.TenantList(tenants, search, status, page, totalPages))
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

	return render(c, admin.TenantDetail(tenant, user, deployments))
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

	user, ok := middleware.GetUser(c)
	if ok {
		tid := id.String()
		LogAction(h.db, user.ID, "tenant.approved", &tid, "", c.IP())
	}

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}
	return c.Redirect("/admin/tenants")
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

	user, ok := middleware.GetUser(c)
	if ok {
		tid := id.String()
		LogAction(h.db, user.ID, "tenant.suspended", &tid, "", c.IP())
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

	user, ok := middleware.GetUser(c)
	if ok {
		tid := id.String()
		LogAction(h.db, user.ID, "provision.retry", &tid, "", c.IP())
	}

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}
	return c.Redirect("/admin/tenants/" + id.String())
}

func (h *Handler) AuditLog(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	const limit = 50
	offset := (page - 1) * limit

	rows, err := h.db.Query(`
		SELECT a.id, u.email, a.action, a.tenant_id, a.detail, a.ip_address, a.created_at
		FROM audit_logs a
		JOIN users u ON u.id = a.admin_id
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(
			&l.ID, &l.AdminEmail, &l.Action,
			&l.TenantID, &l.Detail, &l.IPAddress, &l.CreatedAt,
		); err != nil {
			return err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return render(c, admin.AuditLogPage(logs, page))
}

func (h *Handler) ExportAuditCSV(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="audit_log.csv"`)

	rows, err := h.db.Query(`
		SELECT a.id, u.email, a.action, a.tenant_id, a.detail, a.ip_address, a.created_at
		FROM audit_logs a
		JOIN users u ON u.id = a.admin_id
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := csv.NewWriter(c.Response().BodyWriter())
	w.Write([]string{"ID", "Admin Email", "Action", "Tenant ID", "Detail", "IP", "Created At"})

	for rows.Next() {
		var (
			id                                 int64
			adminEmail, action, detail, ip, ts string
			tenantID                           *string
		)
		if err := rows.Scan(&id, &adminEmail, &action, &tenantID, &detail, &ip, &ts); err != nil {
			return err
		}
		tid := ""
		if tenantID != nil {
			tid = *tenantID
		}
		w.Write([]string{
			strconv.FormatInt(id, 10), adminEmail, action, tid, detail, ip, ts,
		})
	}

	w.Flush()
	return w.Error()
}
