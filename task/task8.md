Task 8 - Dashboard Stats

Objective:
Add admin dashboard stats for tenants, deployments, and recent audit actions.

---

### 1) Model

Create internal/models/dashboard.go:
```go
package models

type DashboardStats struct {
    TotalTenants       int
    ActiveTenants      int
    PendingTenants     int
    SuspendedTenants   int
    RunningDeployments int
    FailedDeployments  int
    RecentActions      []AuditLog
}
```

---

### 2) Handler update

Update AdminDashboard in internal/handlers/admin.go:
```go
func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
    var stats models.DashboardStats

    // Tenant counts
    err := h.db.QueryRow(`
        SELECT
            COUNT(*),
            COALESCE(SUM(CASE WHEN status='active'    THEN 1 ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN status='pending'   THEN 1 ELSE 0 END), 0),
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

    // Deployment counts
    err = h.db.QueryRow(`
        SELECT
            COALESCE(SUM(CASE WHEN status='running' THEN 1 ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN status='failed'  THEN 1 ELSE 0 END), 0)
        FROM deployments
    `).Scan(
        &stats.RunningDeployments,
        &stats.FailedDeployments,
    )
    if err != nil {
        return err
    }

    // Last 10 audit actions
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

    return render(c, views.AdminDashboardPage(stats))
}
```

Note: COALESCE prevents NULL panics when tables are empty.

---

### 3) View update

Update the admin dashboard template to render:
- Total, active, pending, suspended tenants
- Running and failed deployments
- Recent audit actions (last 10)

---

### 4) Acceptance Criteria

1) Dashboard shows correct counts.
2) Recent actions list appears.
3) Page renders with empty tables.

Expected Outcome:
- Admin dashboard surfaces operational stats.

Priority:
MEDIUM
