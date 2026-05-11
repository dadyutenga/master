# HMS Control — Admin Side Full Implementation

> Single role: `admin` — full access to everything.
> Stack: Go + Fiber v2 + templ + SQLite

---

## Table of Contents

1. [Role Simplification](#1-role-simplification)
2. [Audit Log](#2-audit-log)
3. [Admin Dashboard Stats](#3-admin-dashboard-stats)
4. [Tenant List — Search, Filter, Pagination](#4-tenant-list--search-filter-pagination)
5. [Deployment Health Check](#5-deployment-health-check)
6. [SSE Live Provisioning Logs](#6-sse-live-provisioning-logs)
7. [Settings Expansion](#7-settings-expansion)
8. [Tenant Impersonation](#8-tenant-impersonation)
9. [Export CSV](#9-export-csv)
10. [Git Cleanup](#10-git-cleanup)
11. [main.go — Final Route Registration](#11-maingo--final-route-registration)
12. [CSS Additions](#12-css-additions)
13. [Implementation Order](#13-implementation-order)

---

## 1. Role Simplification

### `migrations/005_rename_superadmin.sql`
```sql
UPDATE users SET role = 'admin' WHERE role = 'superadmin';
```

### `internal/middleware/auth.go` — full file
```go
package middleware

import (
    "database/sql"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/session"
)

// Auth validates that a session exists and the user is active.
// Sets "user_id", "user_email", "user_role", and "impersonating" as locals.
func Auth(store *session.Store, db *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess, err := store.Get(c)
        if err != nil {
            return c.Redirect("/login")
        }

        userID, ok := sess.Get("user_id").(int64)
        if !ok || userID == 0 {
            return c.Redirect("/login")
        }

        // Check for impersonation — load impersonated tenant instead of real admin
        if impID, ok := sess.Get("impersonating_tenant_id").(int64); ok && impID != 0 {
            var email, role string
            err := db.QueryRow(
                `SELECT email, role FROM users WHERE id = ? AND status = 'active'`, impID,
            ).Scan(&email, &role)
            if err != nil {
                // Target gone — clean up and redirect
                sess.Delete("impersonating_tenant_id")
                sess.Save()
                return c.Redirect("/admin/tenants")
            }
            c.Locals("user_id", impID)
            c.Locals("user_email", email)
            c.Locals("user_role", role)
            c.Locals("impersonating", true)
            c.Locals("original_admin_id", userID)
            return c.Next()
        }

        var email, role string
        err = db.QueryRow(
            `SELECT email, role FROM users WHERE id = ? AND status = 'active'`, userID,
        ).Scan(&email, &role)
        if err != nil {
            sess.Destroy()
            return c.Redirect("/login")
        }

        c.Locals("user_id", userID)
        c.Locals("user_email", email)
        c.Locals("user_role", role)
        c.Locals("impersonating", false)
        return c.Next()
    }
}

// RequireRole blocks access unless the session user has the given role.
// Also blocks access during impersonation — admin is acting as a tenant.
func RequireRole(role string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        if imp, _ := c.Locals("impersonating").(bool); imp {
            return fiber.ErrForbidden
        }
        if r, _ := c.Locals("user_role").(string); r != role {
            return fiber.ErrForbidden
        }
        return c.Next()
    }
}
```

---

## 2. Audit Log

### `migrations/006_audit_log.sql`
```sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id   INTEGER NOT NULL,
    action     TEXT    NOT NULL,
    tenant_id  INTEGER,
    detail     TEXT,
    ip_address TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id  ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action     ON audit_logs(action);
```

### `internal/models/audit.go`
```go
package models

import (
    "database/sql"
    "time"
)

type AuditLog struct {
    ID         int64
    AdminID    int64
    AdminEmail string   // joined from users
    Action     string
    TenantID   *int64
    TenantName *string  // joined from tenants (nullable)
    Detail     string
    IPAddress  string
    CreatedAt  time.Time
}

type AuditPage struct {
    Logs       []AuditLog
    Page       int
    TotalPages int
    Total      int
}

type AuditStore struct{ db *sql.DB }

func NewAuditStore(db *sql.DB) *AuditStore { return &AuditStore{db: db} }

// Log inserts an audit record. Errors are swallowed — never block
// the real action because of an audit failure.
func (s *AuditStore) Log(adminID int64, action string, tenantID *int64, detail, ip string) {
    s.db.Exec(
        `INSERT INTO audit_logs (admin_id, action, tenant_id, detail, ip_address)
         VALUES (?, ?, ?, ?, ?)`,
        adminID, action, tenantID, detail, ip,
    )
}

func (s *AuditStore) List(page, limit int, action, search string) (AuditPage, error) {
    offset := (page - 1) * limit

    base := `FROM audit_logs a
             LEFT JOIN users   u ON u.id = a.admin_id
             LEFT JOIN tenants t ON t.id = a.tenant_id
             WHERE 1=1`
    args := []interface{}{}

    if action != "" {
        base += " AND a.action = ?"
        args = append(args, action)
    }
    if search != "" {
        base += " AND (u.email LIKE ? OR t.name LIKE ?)"
        args = append(args, "%"+search+"%", "%"+search+"%")
    }

    var total int
    s.db.QueryRow("SELECT COUNT(*) "+base, args...).Scan(&total)

    rows, err := s.db.Query(
        `SELECT a.id, a.admin_id, COALESCE(u.email,'deleted'),
                a.action, a.tenant_id, t.name,
                COALESCE(a.detail,''), COALESCE(a.ip_address,''), a.created_at
         `+base+` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`,
        append(args, limit, offset)...,
    )
    if err != nil {
        return AuditPage{}, err
    }
    defer rows.Close()

    var logs []AuditLog
    for rows.Next() {
        var l AuditLog
        var createdAt string
        rows.Scan(
            &l.ID, &l.AdminID, &l.AdminEmail, &l.Action,
            &l.TenantID, &l.TenantName, &l.Detail, &l.IPAddress, &createdAt,
        )
        l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
        logs = append(logs, l)
    }

    totalPages := (total + limit - 1) / limit
    if totalPages == 0 {
        totalPages = 1
    }
    return AuditPage{Logs: logs, Page: page, TotalPages: totalPages, Total: total}, nil
}
```

### `internal/handlers/audit.go`
```go
package handlers

import (
    "encoding/csv"
    "fmt"
    "strconv"
    "time"

    "github.com/gofiber/fiber/v2"
)

// GET /admin/audit
func (h *Handler) AuditLog(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    action  := c.Query("action", "")
    search  := c.Query("q", "")
    if page < 1 {
        page = 1
    }

    result, err := h.audit.List(page, 50, action, search)
    if err != nil {
        return fmt.Errorf("audit list: %w", err)
    }
    return Render(c, AuditLogPage(result, action, search))
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
        if l.TenantID   != nil { tid   = strconv.FormatInt(*l.TenantID, 10) }
        if l.TenantName != nil { tname = *l.TenantName }
        w.Write([]string{
            strconv.FormatInt(l.ID, 10),
            l.AdminEmail, l.Action, tid, tname,
            l.Detail, l.IPAddress, l.CreatedAt.Format(time.RFC3339),
        })
    }
    w.Flush()
    return w.Error()
}
```

### `internal/views/audit_log.templ`
```go
package views

import (
    "fmt"
    "github.com/dadyutenga/hms-control/internal/models"
)

templ AuditLogPage(result models.AuditPage, action, search string) {
    @AdminLayout("Audit Log") {
        <div class="page-header">
            <h1 class="page-title">Audit Log</h1>
            <a href="/admin/audit/export" class="btn btn-outline">↓ Export CSV</a>
        </div>

        <form method="GET" action="/admin/audit" class="filter-bar">
            <input type="text" name="q" placeholder="Search admin or tenant…"
                   value={ search } class="input" />
            <select name="action" class="input">
                <option value="">All Actions</option>
                for _, a := range []string{
                    "tenant.approved","tenant.suspended","tenant.impersonated",
                    "provision.retry","deployment.start","deployment.stop",
                    "billing.updated","settings.contact","settings.smtp","settings.provisioner",
                } {
                    if a == action {
                        <option value={ a } selected>{ a }</option>
                    } else {
                        <option value={ a }>{ a }</option>
                    }
                }
            </select>
            <button type="submit" class="btn btn-primary">Filter</button>
            if search != "" || action != "" {
                <a href="/admin/audit" class="btn btn-outline">Clear</a>
            }
        </form>

        <p class="text-muted" style="margin-bottom:1rem;">
            { fmt.Sprintf("%d records", result.Total) }
        </p>

        <div class="table-wrap">
            <table class="table">
                <thead>
                    <tr>
                        <th>When</th><th>Admin</th><th>Action</th>
                        <th>Tenant</th><th>Detail</th><th>IP</th>
                    </tr>
                </thead>
                <tbody>
                    for _, l := range result.Logs {
                        <tr>
                            <td class="text-muted text-sm">{ l.CreatedAt.Format("02 Jan 06 15:04") }</td>
                            <td>{ l.AdminEmail }</td>
                            <td><span class={ "badge " + actionBadgeClass(l.Action) }>{ l.Action }</span></td>
                            <td>
                                if l.TenantName != nil {
                                    <a href={ templ.SafeURL("/admin/tenants/" + fmt.Sprint(*l.TenantID)) }>
                                        { *l.TenantName }
                                    </a>
                                } else {
                                    <span class="text-muted">—</span>
                                }
                            </td>
                            <td class="text-sm">{ l.Detail }</td>
                            <td class="text-muted text-sm">{ l.IPAddress }</td>
                        </tr>
                    }
                    if len(result.Logs) == 0 {
                        <tr>
                            <td colspan="6" style="text-align:center;padding:2rem;" class="text-muted">
                                No audit records found.
                            </td>
                        </tr>
                    }
                </tbody>
            </table>
        </div>

        @Pagination(result.Page, result.TotalPages, "/admin/audit",
            map[string]string{"q": search, "action": action})
    }
}

func actionBadgeClass(action string) string {
    switch action {
    case "tenant.approved":      return "badge-green"
    case "tenant.suspended":     return "badge-red"
    case "tenant.impersonated":  return "badge-yellow"
    case "deployment.start":     return "badge-blue"
    case "deployment.stop":      return "badge-orange"
    default:                     return "badge-gray"
    }
}
```

---

## 3. Admin Dashboard Stats

### `internal/models/dashboard.go`
```go
package models

import (
    "database/sql"
)

type DashboardStats struct {
    TotalTenants       int
    ActiveTenants      int
    PendingTenants     int
    SuspendedTenants   int
    RunningDeployments int
    FailedDeployments  int
    RecentActions      []AuditLog
}

func LoadDashboardStats(db *sql.DB, audit *AuditStore) (DashboardStats, error) {
    var s DashboardStats

    err := db.QueryRow(`
        SELECT
            COUNT(*),
            SUM(CASE WHEN status='active'    THEN 1 ELSE 0 END),
            SUM(CASE WHEN status='pending'   THEN 1 ELSE 0 END),
            SUM(CASE WHEN status='suspended' THEN 1 ELSE 0 END)
        FROM tenants
    `).Scan(&s.TotalTenants, &s.ActiveTenants, &s.PendingTenants, &s.SuspendedTenants)
    if err != nil && err != sql.ErrNoRows {
        return s, err
    }

    db.QueryRow(`
        SELECT
            SUM(CASE WHEN status='running' THEN 1 ELSE 0 END),
            SUM(CASE WHEN status='failed'  THEN 1 ELSE 0 END)
        FROM deployments
    `).Scan(&s.RunningDeployments, &s.FailedDeployments)

    page, _ := audit.List(1, 10, "", "")
    s.RecentActions = page.Logs

    return s, nil
}
```

### `internal/handlers/dashboard.go`
```go
package handlers

import (
    "fmt"

    "github.com/gofiber/fiber/v2"
    "github.com/dadyutenga/hms-control/internal/models"
)

// GET /admin/
func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
    stats, err := models.LoadDashboardStats(h.db, h.audit)
    if err != nil {
        return fmt.Errorf("dashboard stats: %w", err)
    }
    return Render(c, AdminDashboardPage(stats))
}
```

### `internal/views/admin_dashboard.templ`
```go
package views

import (
    "fmt"
    "github.com/dadyutenga/hms-control/internal/models"
)

templ AdminDashboardPage(stats models.DashboardStats) {
    @AdminLayout("Dashboard") {
        <h1 class="page-title">Dashboard</h1>

        <div class="stats-grid">
            @StatCard("Total Tenants",   fmt.Sprint(stats.TotalTenants),       "👥")
            @StatCard("Active",          fmt.Sprint(stats.ActiveTenants),       "✅")
            @StatCard("Pending",         fmt.Sprint(stats.PendingTenants),      "⏳")
            @StatCard("Suspended",       fmt.Sprint(stats.SuspendedTenants),    "🚫")
            @StatCard("Running Deploys", fmt.Sprint(stats.RunningDeployments),  "🚀")
            @StatCard("Failed Deploys",  fmt.Sprint(stats.FailedDeployments),   "💥")
        </div>

        <div class="quick-links">
            <a href={ templ.SafeURL("/admin/tenants?status=pending") } class="btn btn-primary">
                Review Pending ({ fmt.Sprint(stats.PendingTenants) })
            </a>
            <a href="/admin/audit"           class="btn btn-outline">Audit Log</a>
            <a href="/admin/tenants/export"  class="btn btn-outline">Export CSV</a>
        </div>

        <section style="margin-top:2rem;">
            <h2 class="section-title">Recent Activity</h2>
            <div class="table-wrap">
                <table class="table">
                    <thead>
                        <tr><th>When</th><th>Admin</th><th>Action</th><th>Tenant</th></tr>
                    </thead>
                    <tbody>
                        for _, l := range stats.RecentActions {
                            <tr>
                                <td class="text-muted text-sm">{ l.CreatedAt.Format("02 Jan 15:04") }</td>
                                <td>{ l.AdminEmail }</td>
                                <td>
                                    <span class={ "badge " + actionBadgeClass(l.Action) }>
                                        { l.Action }
                                    </span>
                                </td>
                                <td>
                                    if l.TenantName != nil {
                                        { *l.TenantName }
                                    } else {
                                        <span class="text-muted">—</span>
                                    }
                                </td>
                            </tr>
                        }
                    </tbody>
                </table>
            </div>
        </section>
    }
}

templ StatCard(label, value, icon string) {
    <div class="stat-card">
        <div class="stat-icon">{ icon }</div>
        <div class="stat-value">{ value }</div>
        <div class="stat-label">{ label }</div>
    </div>
}
```

---

## 4. Tenant List — Search, Filter, Pagination

### `internal/models/tenant_list.go`
```go
package models

import (
    "database/sql"
    "time"
)

type TenantRow struct {
    ID        int64
    Name      string
    Email     string
    Status    string
    Plan      string
    CreatedAt time.Time
}

type TenantPage struct {
    Tenants    []TenantRow
    Page       int
    TotalPages int
    Total      int
    Status     string
    Search     string
}

func ListTenants(db *sql.DB, page, limit int, status, search string) (TenantPage, error) {
    offset := (page - 1) * limit

    base := "FROM tenants WHERE 1=1"
    args := []interface{}{}

    if status != "" {
        base += " AND status = ?"
        args = append(args, status)
    }
    if search != "" {
        base += " AND (name LIKE ? OR email LIKE ?)"
        args = append(args, "%"+search+"%", "%"+search+"%")
    }

    var total int
    db.QueryRow("SELECT COUNT(*) "+base, args...).Scan(&total)

    rows, err := db.Query(
        "SELECT id, name, email, status, COALESCE(plan,'free'), created_at "+
            base+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
        append(args, limit, offset)...,
    )
    if err != nil {
        return TenantPage{}, err
    }
    defer rows.Close()

    var tenants []TenantRow
    for rows.Next() {
        var t TenantRow
        var createdAt string
        rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status, &t.Plan, &createdAt)
        t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
        tenants = append(tenants, t)
    }

    totalPages := (total + limit - 1) / limit
    if totalPages == 0 {
        totalPages = 1
    }
    return TenantPage{
        Tenants: tenants, Page: page, TotalPages: totalPages,
        Total: total, Status: status, Search: search,
    }, nil
}
```

### `internal/handlers/tenants.go` — update ListTenants
```go
// GET /admin/tenants
func (h *Handler) ListTenants(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    status  := c.Query("status", "")
    search  := c.Query("q", "")
    if page < 1 {
        page = 1
    }

    result, err := models.ListTenants(h.db, page, 20, status, search)
    if err != nil {
        return fmt.Errorf("list tenants: %w", err)
    }
    return Render(c, TenantListPage(result))
}
```

### `internal/views/tenant_list.templ`
```go
package views

import (
    "fmt"
    "github.com/dadyutenga/hms-control/internal/models"
)

templ TenantListPage(result models.TenantPage) {
    @AdminLayout("Tenants") {
        <div class="page-header">
            <h1 class="page-title">
                Tenants
                <span class="badge badge-gray">{ fmt.Sprint(result.Total) }</span>
            </h1>
            <a href="/admin/tenants/export" class="btn btn-outline">↓ Export CSV</a>
        </div>

        <form method="GET" action="/admin/tenants" class="filter-bar">
            <input type="text" name="q" placeholder="Search name or email…"
                   value={ result.Search } class="input" />
            <select name="status" class="input">
                <option value="">All</option>
                for _, s := range []string{"pending","active","suspended"} {
                    if s == result.Status {
                        <option value={ s } selected>{ s }</option>
                    } else {
                        <option value={ s }>{ s }</option>
                    }
                }
            </select>
            <button type="submit" class="btn btn-primary">Filter</button>
            if result.Search != "" || result.Status != "" {
                <a href="/admin/tenants" class="btn btn-outline">Clear</a>
            }
        </form>

        <div class="table-wrap">
            <table class="table">
                <thead>
                    <tr>
                        <th>Name</th><th>Email</th><th>Plan</th>
                        <th>Status</th><th>Joined</th><th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    for _, t := range result.Tenants {
                        <tr>
                            <td>
                                <a href={ templ.SafeURL("/admin/tenants/" + fmt.Sprint(t.ID)) }
                                   class="font-medium">{ t.Name }</a>
                            </td>
                            <td class="text-muted">{ t.Email }</td>
                            <td><span class="badge badge-gray">{ t.Plan }</span></td>
                            <td><span class={ "badge " + StatusBadgeClass(t.Status) }>{ t.Status }</span></td>
                            <td class="text-muted text-sm">{ t.CreatedAt.Format("02 Jan 2006") }</td>
                            <td>
                                <div class="action-row">
                                    <a href={ templ.SafeURL("/admin/tenants/" + fmt.Sprint(t.ID)) }
                                       class="btn-xs btn-outline">View</a>
                                    if t.Status == "pending" {
                                        <form method="POST" action={ templ.SafeURL(fmt.Sprintf("/admin/tenants/%d/approve", t.ID)) }>
                                            <button class="btn-xs btn-green">Approve</button>
                                        </form>
                                    }
                                    if t.Status == "active" {
                                        <form method="POST" action={ templ.SafeURL(fmt.Sprintf("/admin/tenants/%d/suspend", t.ID)) }
                                              onsubmit="return confirm('Suspend this tenant?')">
                                            <button class="btn-xs btn-red">Suspend</button>
                                        </form>
                                    }
                                </div>
                            </td>
                        </tr>
                    }
                    if len(result.Tenants) == 0 {
                        <tr>
                            <td colspan="6" style="text-align:center;padding:2rem;" class="text-muted">
                                No tenants found.
                            </td>
                        </tr>
                    }
                </tbody>
            </table>
        </div>

        @Pagination(result.Page, result.TotalPages, "/admin/tenants",
            map[string]string{"q": result.Search, "status": result.Status})
    }
}

func StatusBadgeClass(status string) string {
    switch status {
    case "active":    return "badge-green"
    case "pending":   return "badge-yellow"
    case "suspended": return "badge-red"
    default:          return "badge-gray"
    }
}
```

---

## 5. Deployment Health Check

### `internal/handlers/health.go`
```go
package handlers

import (
    "database/sql"
    "net/http"
    "time"

    "github.com/gofiber/fiber/v2"
)

type HealthResult struct {
    Status     string `json:"status"`       // "UP" | "DOWN" | "UNKNOWN"
    TenantID   string `json:"tenant_id"`
    Endpoint   string `json:"endpoint"`
    StatusCode int    `json:"status_code"`
    LatencyMs  int64  `json:"latency_ms"`
    Error      string `json:"error,omitempty"`
}

// GET /admin/tenants/:id/health
func (h *Handler) TenantHealthCheck(c *fiber.Ctx) error {
    tenantID := c.Params("id")

    var endpoint string
    err := h.db.QueryRow(
        `SELECT endpoint FROM deployments
         WHERE tenant_id = ? AND status = 'running'
         ORDER BY created_at DESC LIMIT 1`,
        tenantID,
    ).Scan(&endpoint)

    if err == sql.ErrNoRows {
        return c.JSON(HealthResult{
            Status:   "UNKNOWN",
            TenantID: tenantID,
            Error:    "no running deployment found",
        })
    }
    if err != nil {
        return c.JSON(HealthResult{Status: "UNKNOWN", TenantID: tenantID, Error: err.Error()})
    }

    client := &http.Client{Timeout: 5 * time.Second}
    start  := time.Now()
    resp, reqErr := client.Get(endpoint + "/health")
    latency := time.Since(start).Milliseconds()

    if reqErr != nil {
        return c.JSON(HealthResult{
            Status: "DOWN", TenantID: tenantID,
            Endpoint: endpoint, LatencyMs: latency,
            Error: reqErr.Error(),
        })
    }
    resp.Body.Close()

    status := "DOWN"
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        status = "UP"
    }
    return c.JSON(HealthResult{
        Status: status, TenantID: tenantID,
        Endpoint: endpoint, StatusCode: resp.StatusCode, LatencyMs: latency,
    })
}
```

### Snippet — add to `show_tenant.templ` next to deployment status
```go
<!-- Health check badge — auto-refreshes every 30s -->
<span id="health-badge" class="badge badge-gray">Checking…</span>
<span id="health-latency" class="text-muted text-sm"></span>

<script>
  (function() {
    const badge   = document.getElementById('health-badge');
    const latency = document.getElementById('health-latency');

    function check() {
      fetch('/admin/tenants/TENANT_ID/health')
        .then(r => r.json())
        .then(d => {
          badge.textContent = d.status;
          badge.className   = 'badge ' + (d.status === 'UP' ? 'badge-green' : 'badge-red');
          if (d.latency_ms) latency.textContent = d.latency_ms + 'ms';
          if (d.error)      latency.textContent = d.error;
        })
        .catch(() => { badge.textContent = 'ERR'; badge.className = 'badge badge-red'; });
    }

    check();
    setInterval(check, 30000);
  })();
</script>
```
> Replace `TENANT_ID` with the templ expression: `{ fmt.Sprint(tenant.ID) }`

---

## 6. SSE Live Provisioning Logs

### `internal/handlers/logs.go`
```go
package handlers

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
)

// GET /admin/tenants/:id/logs/stream
// Streams per-tenant provisioning log as Server-Sent Events.
func (h *Handler) StreamProvisionLogs(c *fiber.Ctx) error {
    tenantID := c.Params("id")
    logPath  := fmt.Sprintf("./tmp/provision-%s.log", tenantID)

    c.Set("Content-Type",       "text/event-stream")
    c.Set("Cache-Control",      "no-cache")
    c.Set("Connection",         "keep-alive")
    c.Set("X-Accel-Buffering",  "no") // disable nginx buffering

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
                // SSE comment keeps connection alive through proxies
                fmt.Fprintf(w, ": heartbeat\n\n")
                w.Flush()
            default:
                line, err := reader.ReadString('\n')
                if len(line) > 0 {
                    line = strings.TrimRight(line, "\r\n")
                    fmt.Fprintf(w, "data: %s\n\n", line)
                    w.Flush()
                }
                if err == io.EOF {
                    // Check if provisioning is done
                    var provStatus string
                    h.db.QueryRow(
                        `SELECT status FROM deployments WHERE tenant_id = ?
                         ORDER BY created_at DESC LIMIT 1`, tenantID,
                    ).Scan(&provStatus)

                    if provStatus == "running" || provStatus == "failed" {
                        fmt.Fprintf(w, "event: done\ndata: %s\n\n", provStatus)
                        w.Flush()
                        return
                    }
                    time.Sleep(500 * time.Millisecond)
                    continue
                }
                if err != nil {
                    fmt.Fprintf(w, "data: [stream error] %s\n\n", err.Error())
                    w.Flush()
                    return
                }
            }
        }
    })

    return nil
}
```

### Templ component — `internal/views/log_viewer.templ`
```go
package views

import "fmt"

templ LogViewer(tenantID int64) {
    <section class="log-section" style="margin-top:2rem;">
        <div class="log-header">
            <h2 class="section-title">Provisioning Log</h2>
            <div style="display:flex;gap:0.5rem;align-items:center;">
                <button id="log-clear"  class="btn-xs btn-outline">Clear</button>
                <button id="log-pause"  class="btn-xs btn-outline">Pause</button>
                <span   id="log-status" class="badge badge-gray">Connecting…</span>
            </div>
        </div>
        <pre id="log-output"
             style="height:400px;overflow-y:auto;background:#0d1117;color:#c9d1d9;
                    padding:1rem;border-radius:0.5rem;font-size:0.8rem;
                    line-height:1.5;font-family:monospace;white-space:pre-wrap;">
        </pre>
    </section>

    <script>
      (function() {
        const out      = document.getElementById('log-output');
        const badge    = document.getElementById('log-status');
        const tid      = '{ fmt.Sprint(tenantID) }';
        const MAX_LINES = 500;
        let es          = null;
        let paused      = false;
        let lineCount   = 0;

        function connect() {
          es = new EventSource('/admin/tenants/' + tid + '/logs/stream');

          es.onopen = () => {
            badge.textContent = 'Live';
            badge.className   = 'badge badge-green';
          };

          es.onmessage = e => {
            if (paused) return;
            lineCount++;
            out.textContent += e.data + '\n';
            if (lineCount > MAX_LINES) {
              const lines = out.textContent.split('\n');
              out.textContent = lines.slice(-MAX_LINES).join('\n');
              lineCount = MAX_LINES;
            }
            out.scrollTop = out.scrollHeight;
          };

          es.addEventListener('done', e => {
            const ok = e.data === 'running';
            badge.textContent = ok ? '✅ Done' : '❌ Failed';
            badge.className   = ok ? 'badge badge-green' : 'badge badge-red';
            es.close();
          });

          es.onerror = () => {
            badge.textContent = 'Disconnected';
            badge.className   = 'badge badge-red';
            es.close();
            setTimeout(connect, 3000); // auto-reconnect
          };
        }

        document.getElementById('log-clear').onclick = () => {
          out.textContent = ''; lineCount = 0;
        };
        document.getElementById('log-pause').onclick = function() {
          paused = !paused;
          this.textContent = paused ? 'Resume' : 'Pause';
        };

        connect();
      })();
    </script>
}
```

---

## 7. Settings Expansion

### `migrations/007_settings.sql`
```sql
CREATE TABLE IF NOT EXISTS settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO settings (key, value) VALUES
    ('smtp_host',         ''),
    ('smtp_port',         '587'),
    ('smtp_user',         ''),
    ('smtp_pass',         ''),
    ('smtp_from',         'noreply@localhost'),
    ('smtp_tls',          'true'),
    ('provision_script',  './scripts/provision.sh'),
    ('docker_template',   'default'),
    ('provision_timeout', '300');
```

### `internal/models/settings.go`
```go
package models

import "database/sql"

type SettingsStore struct{ db *sql.DB }

func NewSettingsStore(db *sql.DB) *SettingsStore { return &SettingsStore{db: db} }

func (s *SettingsStore) Get(key string) string {
    var val string
    s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&val)
    return val
}

func (s *SettingsStore) Set(key, value string) error {
    _, err := s.db.Exec(`
        INSERT INTO settings (key, value, updated_at)
        VALUES (?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(key) DO UPDATE
            SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
        key, value,
    )
    return err
}

func (s *SettingsStore) All() (map[string]string, error) {
    rows, err := s.db.Query(`SELECT key, value FROM settings`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    m := make(map[string]string)
    for rows.Next() {
        var k, v string
        rows.Scan(&k, &v)
        m[k] = v
    }
    return m, nil
}
```

### `internal/handlers/settings.go`
```go
package handlers

import (
    "fmt"
    "net/smtp"

    "github.com/gofiber/fiber/v2"
)

// GET /admin/settings/smtp
func (h *Handler) AdminSMTPSettings(c *fiber.Ctx) error {
    s, err := h.settings.All()
    if err != nil {
        return err
    }
    saved := c.Query("saved") == "1"
    return Render(c, SMTPSettingsPage(s, saved))
}

// POST /admin/settings/smtp
func (h *Handler) UpdateSMTPSettings(c *fiber.Ctx) error {
    adminID := c.Locals("user_id").(int64)

    fields := map[string]string{
        "smtp_host": c.FormValue("smtp_host"),
        "smtp_port": c.FormValue("smtp_port"),
        "smtp_user": c.FormValue("smtp_user"),
        "smtp_from": c.FormValue("smtp_from"),
        "smtp_tls":  c.FormValue("smtp_tls"),
    }
    if pass := c.FormValue("smtp_pass"); pass != "" {
        fields["smtp_pass"] = pass // only update if provided
    }
    for k, v := range fields {
        if err := h.settings.Set(k, v); err != nil {
            return fmt.Errorf("save setting %s: %w", k, err)
        }
    }

    h.audit.Log(adminID, "settings.smtp", nil, "SMTP settings updated", c.IP())
    return c.Redirect("/admin/settings/smtp?saved=1")
}

// POST /admin/settings/smtp/test — returns JSON
func (h *Handler) TestSMTP(c *fiber.Ctx) error {
    to   := c.FormValue("test_email")
    host := h.settings.Get("smtp_host")
    port := h.settings.Get("smtp_port")
    user := h.settings.Get("smtp_user")
    pass := h.settings.Get("smtp_pass")
    from := h.settings.Get("smtp_from")

    if host == "" {
        return c.JSON(fiber.Map{"ok": false, "error": "SMTP host not configured"})
    }
    if to == "" {
        return c.JSON(fiber.Map{"ok": false, "error": "provide a test email address"})
    }

    auth := smtp.PlainAuth("", user, pass, host)
    msg  := fmt.Sprintf(
        "From: %s\r\nTo: %s\r\nSubject: HMS SMTP Test\r\n\r\nSMTP is working correctly.",
        from, to,
    )
    err := smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))
    if err != nil {
        return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
    }
    return c.JSON(fiber.Map{"ok": true, "message": "Test email sent to " + to})
}

// GET /admin/settings/provisioner
func (h *Handler) AdminProvisionerSettings(c *fiber.Ctx) error {
    s, err := h.settings.All()
    if err != nil {
        return err
    }
    saved := c.Query("saved") == "1"
    return Render(c, ProvisionerSettingsPage(s, saved))
}

// POST /admin/settings/provisioner
func (h *Handler) UpdateProvisionerSettings(c *fiber.Ctx) error {
    adminID := c.Locals("user_id").(int64)

    for _, k := range []string{"provision_script", "docker_template", "provision_timeout"} {
        if err := h.settings.Set(k, c.FormValue(k)); err != nil {
            return fmt.Errorf("save setting %s: %w", k, err)
        }
    }

    h.audit.Log(adminID, "settings.provisioner", nil, "Provisioner settings updated", c.IP())
    return c.Redirect("/admin/settings/provisioner?saved=1")
}
```

### `internal/views/smtp_settings.templ`
```go
package views

templ SMTPSettingsPage(s map[string]string, saved bool) {
    @AdminLayout("SMTP Settings") {
        <h1 class="page-title">SMTP Settings</h1>

        if saved {
            <div class="alert alert-green">✅ Settings saved successfully.</div>
        }

        <form method="POST" action="/admin/settings/smtp" class="settings-form">
            @SettingField("Host",          "smtp_host", s["smtp_host"], "mail.example.com",    "text")
            @SettingField("Port",          "smtp_port", s["smtp_port"], "587",                 "text")
            @SettingField("Username",      "smtp_user", s["smtp_user"], "user@example.com",    "text")
            @SettingField("Password",      "smtp_pass", "",             "leave blank to keep current", "password")
            @SettingField("From Address",  "smtp_from", s["smtp_from"], "noreply@example.com", "email")
            <div class="form-group">
                <label class="label">TLS</label>
                <select name="smtp_tls" class="input">
                    if s["smtp_tls"] == "true" {
                        <option value="true"  selected>Enabled</option>
                        <option value="false">Disabled</option>
                    } else {
                        <option value="true">Enabled</option>
                        <option value="false" selected>Disabled</option>
                    }
                </select>
            </div>
            <div class="btn-row">
                <button type="submit" class="btn btn-primary">Save Settings</button>
            </div>
        </form>

        <section style="margin-top:2rem;max-width:480px;">
            <h2 class="section-title">Test Connection</h2>
            <div class="filter-bar">
                <input id="test-email" type="email" class="input"
                       placeholder="you@example.com" style="flex:1;" />
                <button onclick="testSMTP()" class="btn btn-outline">Send Test</button>
            </div>
            <p id="test-result" class="text-sm" style="margin-top:0.5rem;"></p>
        </section>

        <script>
          function testSMTP() {
            const to     = document.getElementById('test-email').value.trim();
            const result = document.getElementById('test-result');
            if (!to) { result.textContent = 'Enter an email address first.'; return; }
            result.textContent = 'Sending…';
            const body = new FormData();
            body.append('test_email', to);
            fetch('/admin/settings/smtp/test', { method: 'POST', body })
              .then(r => r.json())
              .then(d => {
                result.textContent = d.ok ? '✅ ' + d.message : '❌ ' + d.error;
                result.style.color = d.ok ? '#065f46' : '#991b1b';
              })
              .catch(() => { result.textContent = '❌ Request failed'; result.style.color='#991b1b'; });
          }
        </script>
    }
}

templ SettingField(label, name, value, placeholder, inputType string) {
    <div class="form-group">
        <label class="label" for={ name }>{ label }</label>
        <input id={ name } type={ inputType } name={ name }
               value={ value } placeholder={ placeholder } class="input" />
    </div>
}
```

### `internal/views/provisioner_settings.templ`
```go
package views

templ ProvisionerSettingsPage(s map[string]string, saved bool) {
    @AdminLayout("Provisioner Settings") {
        <h1 class="page-title">Provisioner Settings</h1>

        if saved {
            <div class="alert alert-green">✅ Settings saved successfully.</div>
        }

        <form method="POST" action="/admin/settings/provisioner" class="settings-form">
            @SettingField("Provision Script Path", "provision_script",
                s["provision_script"], "./scripts/provision.sh", "text")
            @SettingField("Docker Template Name",  "docker_template",
                s["docker_template"],  "default", "text")
            @SettingField("Provision Timeout (seconds)", "provision_timeout",
                s["provision_timeout"], "300", "number")

            <p class="text-muted text-sm" style="margin-top:0.5rem;">
                Script receives: <code>TENANT_ID</code>, <code>TENANT_DOMAIN</code>,
                <code>DOCKER_TEMPLATE</code> as env vars.
            </p>

            <div class="btn-row">
                <button type="submit" class="btn btn-primary">Save Settings</button>
            </div>
        </form>
    }
}
```

---

## 8. Tenant Impersonation

### `internal/handlers/impersonate.go`
```go
package handlers

import (
    "database/sql"
    "fmt"
    "strconv"

    "github.com/gofiber/fiber/v2"
)

// POST /admin/tenants/:id/impersonate
func (h *Handler) ImpersonateTenant(c *fiber.Ctx) error {
    tenantID, err := strconv.ParseInt(c.Params("id"), 10, 64)
    if err != nil {
        return fiber.ErrBadRequest
    }

    // Verify tenant exists and is active
    var email string
    err = h.db.QueryRow(
        `SELECT email FROM users WHERE id = ? AND status = 'active'`, tenantID,
    ).Scan(&email)
    if err == sql.ErrNoRows {
        return c.Status(fiber.StatusNotFound).SendString("tenant user not found or not active")
    }
    if err != nil {
        return fmt.Errorf("impersonate lookup: %w", err)
    }

    sess, err := h.store.Get(c)
    if err != nil {
        return err
    }

    adminID := c.Locals("user_id").(int64)

    sess.Set("original_admin_id",      adminID)
    sess.Set("impersonating_tenant_id", tenantID)
    if err := sess.Save(); err != nil {
        return err
    }

    // Always audit impersonation
    h.audit.Log(adminID, "tenant.impersonated", &tenantID,
        fmt.Sprintf("impersonating %s", email), c.IP())

    return c.Redirect("/dashboard")
}

// POST /impersonate/stop
func (h *Handler) StopImpersonation(c *fiber.Ctx) error {
    sess, err := h.store.Get(c)
    if err != nil {
        return err
    }

    adminID, _ := sess.Get("original_admin_id").(int64)

    sess.Delete("impersonating_tenant_id")
    sess.Delete("original_admin_id")
    sess.Set("user_id", adminID) // restore real admin user_id
    if err := sess.Save(); err != nil {
        return err
    }

    return c.Redirect("/admin/tenants")
}
```

### Impersonation banner — add to `client_layout.templ`
```go
// At the very top of the body, before nav
if isImpersonating {
    <div style="background:#f59e0b;color:#000;padding:0.5rem 1rem;
                text-align:center;font-size:0.875rem;font-weight:600;position:sticky;top:0;z-index:100;">
        ⚠️ Viewing as { tenantEmail } —
        <form method="POST" action="/impersonate/stop" style="display:inline;">
            <button type="submit"
                    style="background:none;border:none;text-decoration:underline;
                           cursor:pointer;font-weight:700;font-size:inherit;">
                Exit Impersonation
            </button>
        </form>
    </div>
}
```

### Impersonate button — add inside `show_tenant.templ`
```go
if tenant.Status == "active" {
    <form method="POST"
          action={ templ.SafeURL(fmt.Sprintf("/admin/tenants/%d/impersonate", tenant.ID)) }
          onsubmit="return confirm('You will be redirected to the tenant dashboard. Continue?')">
        <button class="btn btn-outline">👤 Impersonate Tenant</button>
    </form>
}
```

---

## 9. Export CSV

### `internal/handlers/export.go`
```go
package handlers

import (
    "encoding/csv"
    "fmt"
    "time"

    "github.com/gofiber/fiber/v2"
)

// GET /admin/tenants/export
func (h *Handler) ExportTenantsCSV(c *fiber.Ctx) error {
    c.Set("Content-Type", "text/csv; charset=utf-8")
    c.Set("Content-Disposition",
        fmt.Sprintf(`attachment; filename="tenants-%s.csv"`, time.Now().Format("2006-01-02")))

    rows, err := h.db.Query(`
        SELECT
            t.id,
            t.name,
            t.email,
            t.status,
            COALESCE(t.plan, 'free'),
            COALESCE(d.status, 'none')     AS deploy_status,
            COALESCE(d.endpoint, '')        AS endpoint,
            t.created_at
        FROM tenants t
        LEFT JOIN deployments d ON d.tenant_id = t.id
            AND d.id = (
                SELECT id FROM deployments
                WHERE tenant_id = t.id
                ORDER BY created_at DESC LIMIT 1
            )
        ORDER BY t.created_at DESC
    `)
    if err != nil {
        return err
    }
    defer rows.Close()

    w := csv.NewWriter(c.Response().BodyWriter())
    w.Write([]string{
        "ID", "Name", "Email", "Status", "Plan",
        "Deployment Status", "Endpoint", "Registered At",
    })

    for rows.Next() {
        var id, name, email, status, plan, deployStatus, endpoint, createdAt string
        rows.Scan(&id, &name, &email, &status, &plan, &deployStatus, &endpoint, &createdAt)
        w.Write([]string{id, name, email, status, plan, deployStatus, endpoint, createdAt})
    }

    w.Flush()
    return w.Error()
}
```

---

## 10. Git Cleanup

```bash
# 1. Update .gitignore
cat >> .gitignore << 'EOF'

# Build binaries
hms-control
hms-control.exe

# Logs
hms.log
tmp/*.log

# Air build cache
tmp/

# User uploads (keep folder structure, not files)
uploads/*
!uploads/.gitkeep
EOF

# 2. Remove already-tracked files
git rm --cached hms-control hms-control.exe hms.log 2>/dev/null || true
git rm -r --cached tmp/ 2>/dev/null || true

# 3. Placeholder so uploads/ folder exists on clone
touch uploads/.gitkeep

# 4. Commit
git add .gitignore uploads/.gitkeep
git commit -m "chore: remove tracked binaries, logs, and tmp"
```

---

## 11. `main.go` — Final Route Registration

```go
package main

import (
    "log"

    "github.com/dadyutenga/hms-control/internal/config"
    "github.com/dadyutenga/hms-control/internal/db"
    "github.com/dadyutenga/hms-control/internal/handlers"
    "github.com/dadyutenga/hms-control/internal/mailer"
    "github.com/dadyutenga/hms-control/internal/middleware"
    "github.com/dadyutenga/hms-control/internal/provisioner"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/gofiber/fiber/v2/middleware/session"
)

func main() {
    cfg      := config.Load()
    database := db.Connect(cfg.DBPath)
    defer database.Close()

    mail := mailer.New(cfg)
    eng  := provisioner.NewEngine(cfg, database, mail)
    eng.Start()

    store := session.New(session.Config{
        KeyLookup:    "cookie:hms_session",
        CookieSecure: cfg.CookieSecure,
    })

    app := fiber.New(fiber.Config{ErrorHandler: handlers.ErrorHandler})
    app.Use(logger.New())
    app.Use(recover.New())
    app.Static("/static", "./static")
    app.Static("/img", "./public/img")

    h := handlers.New(cfg, database, mail, store, eng)

    // ── Public ──────────────────────────────────────────────────────────
    app.Get("/",               h.Home)
    app.Get("/about",          h.About)
    app.Get("/contact",        h.Contact)
    app.Get("/register",       h.ShowRegister)
    app.Post("/register/step1",h.RegisterStep1)
    app.Get("/register/step2", h.ShowRegisterStep2)
    app.Post("/register/step2",h.RegisterStep2)
    app.Get("/register/step3", h.ShowRegisterStep3)
    app.Post("/register/step3",h.RegisterStep3)
    app.Get("/register/success",h.ShowRegisterSuccess)
    app.Get("/verify/:token",  h.VerifyEmail)
    app.Get("/login",          h.ShowLogin)
    app.Post("/login",         h.Login)
    app.Post("/logout",        h.Logout)
    app.Get("/verify-notice",  func(c *fiber.Ctx) error {
        return c.SendString("Check your email for a verification link.")
    })

    // ── Stop impersonation (accessible while acting as tenant) ───────────
    app.Post("/impersonate/stop", middleware.Auth(store, database), h.StopImpersonation)

    // ── Client dashboard ─────────────────────────────────────────────────
    client := app.Group("/dashboard", middleware.Auth(store, database))
    client.Get("/",               h.ClientDashboard)
    client.Get("/details",        h.ShowTenantDetails)
    client.Post("/details",       h.UpdateTenantDetails)
    client.Get("/change-password",h.ShowChangePassword)
    client.Post("/change-password",h.ChangePassword)

    // ── Admin ─────────────────────────────────────────────────────────────
    admin := app.Group("/admin",
        middleware.Auth(store, database),
        middleware.RequireRole("admin"),
    )

    // Dashboard
    admin.Get("/", h.AdminDashboard)

    // Tenants — export BEFORE :id to avoid route conflict
    admin.Get("/tenants",                        h.ListTenants)
    admin.Get("/tenants/export",                 h.ExportTenantsCSV)
    admin.Get("/tenants/:id",                    h.ShowTenant)
    admin.Get("/tenants/:id/health",             h.TenantHealthCheck)
    admin.Get("/tenants/:id/logs/stream",        h.StreamProvisionLogs)
    admin.Get("/tenants/:id/deployments/:dId",   h.ShowDeployment)
    admin.Post("/tenants/:id/approve",           h.ApproveTenant)
    admin.Post("/tenants/:id/suspend",           h.SuspendTenant)
    admin.Post("/tenants/:id/retry",             h.RetryProvision)
    admin.Post("/tenants/:id/deployments/start", h.StartTenantDeployment)
    admin.Post("/tenants/:id/deployments/stop",  h.StopTenantDeployment)
    admin.Post("/tenants/:id/billing",           h.UpdateTenantBilling)
    admin.Post("/tenants/:id/impersonate",       h.ImpersonateTenant)

    // Audit log
    admin.Get("/audit",        h.AuditLog)
    admin.Get("/audit/export", h.ExportAuditCSV)

    // Settings
    admin.Get("/settings/contact",           h.AdminContactSettings)
    admin.Post("/settings/contact",          h.UpdateContactSettings)
    admin.Get("/settings/smtp",              h.AdminSMTPSettings)
    admin.Post("/settings/smtp",             h.UpdateSMTPSettings)
    admin.Post("/settings/smtp/test",        h.TestSMTP)
    admin.Get("/settings/provisioner",       h.AdminProvisionerSettings)
    admin.Post("/settings/provisioner",      h.UpdateProvisionerSettings)

    log.Fatal(app.Listen(":8080"))
}
```

### `internal/handlers/handler.go` — updated struct
```go
package handlers

import (
    "database/sql"

    "github.com/dadyutenga/hms-control/internal/config"
    "github.com/dadyutenga/hms-control/internal/mailer"
    "github.com/dadyutenga/hms-control/internal/models"
    "github.com/dadyutenga/hms-control/internal/provisioner"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/session"
)

type Handler struct {
    cfg      *config.Config
    db       *sql.DB
    mail     *mailer.Mailer
    store    *session.Store
    eng      *provisioner.Engine
    audit    *models.AuditStore    // ← new
    settings *models.SettingsStore // ← new
}

func New(
    cfg *config.Config,
    db *sql.DB,
    mail *mailer.Mailer,
    store *session.Store,
    eng *provisioner.Engine,
) *Handler {
    return &Handler{
        cfg:      cfg,
        db:       db,
        mail:     mail,
        store:    store,
        eng:      eng,
        audit:    models.NewAuditStore(db),
        settings: models.NewSettingsStore(db),
    }
}

// Render is a helper so handlers don't repeat component boilerplate
func Render(c *fiber.Ctx, component interface{ Render(ctx interface{}, w interface{}) error }) error {
    c.Set("Content-Type", "text/html; charset=utf-8")
    return component.Render(c.Context(), c.Response().BodyWriter())
}

// ErrorHandler is used in fiber.Config
func ErrorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
    }
    return c.Status(code).SendString(err.Error())
}
```

### Shared templ components — `internal/views/shared.templ`
```go
package views

import "fmt"

// Pagination renders prev/next links preserving existing query params.
templ Pagination(page, totalPages int, base string, params map[string]string) {
    if totalPages > 1 {
        <div class="pagination">
            if page > 1 {
                <a href={ templ.SafeURL(pageURL(base, page-1, params)) }
                   class="btn btn-outline btn-sm">← Prev</a>
            }
            <span class="text-muted text-sm">
                { fmt.Sprintf("Page %d of %d", page, totalPages) }
            </span>
            if page < totalPages {
                <a href={ templ.SafeURL(pageURL(base, page+1, params)) }
                   class="btn btn-outline btn-sm">Next →</a>
            }
        </div>
    }
}

func pageURL(base string, page int, params map[string]string) string {
    q := fmt.Sprintf("?page=%d", page)
    for k, v := range params {
        if v != "" {
            q += fmt.Sprintf("&%s=%s", k, v)
        }
    }
    return base + q
}
```

---

## 12. CSS Additions — `static/css/admin.css`

```css
/* ── Stat grid ─────────────────────────────────────── */
.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
}
.stat-card {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 1rem;
    padding: 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}
.stat-icon  { font-size: 1.5rem; }
.stat-value { font-size: 2rem; font-weight: 700; line-height: 1; }
.stat-label { font-size: 0.75rem; color: var(--color-muted-fg); }

/* ── Table ─────────────────────────────────────────── */
.table-wrap { overflow-x: auto; border-radius: 0.75rem; border: 1px solid var(--color-border); }
.table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.table th {
    text-align: left;
    padding: 0.625rem 0.875rem;
    border-bottom: 2px solid var(--color-border);
    font-weight: 600;
    color: var(--color-muted-fg);
    white-space: nowrap;
    background: var(--color-elevated);
}
.table td { padding: 0.625rem 0.875rem; border-bottom: 1px solid var(--color-border); vertical-align: middle; }
.table tr:last-child td { border-bottom: none; }
.table tr:hover td { background: var(--color-elevated); }

/* ── Badges ────────────────────────────────────────── */
.badge        { display:inline-flex;align-items:center;padding:0.125rem 0.5rem;border-radius:9999px;font-size:0.7rem;font-weight:600;white-space:nowrap; }
.badge-green  { background:#d1fae5;color:#065f46; }
.badge-red    { background:#fee2e2;color:#991b1b; }
.badge-yellow { background:#fef3c7;color:#92400e; }
.badge-blue   { background:#dbeafe;color:#1e40af; }
.badge-orange { background:#ffedd5;color:#9a3412; }
.badge-gray   { background:var(--color-secondary);color:var(--color-muted-fg); }

/* ── Inputs ────────────────────────────────────────── */
.input {
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--color-border);
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-family: var(--font-body);
    background: var(--color-bg);
    color: var(--color-fg);
    outline: none;
    transition: border-color 0.15s;
}
.input:focus { border-color: var(--color-fg); }

/* ── Filter bar ────────────────────────────────────── */
.filter-bar {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    flex-wrap: wrap;
    margin-bottom: 1.5rem;
}

/* ── Page header ───────────────────────────────────── */
.page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
    gap: 0.75rem;
}
.page-title  { font-size: 1.5rem; font-weight: 700; letter-spacing: -0.025em; }
.section-title { font-size: 1.125rem; font-weight: 600; }

/* ── Forms ─────────────────────────────────────────── */
.settings-form { max-width: 480px; }
.form-group    { display:flex;flex-direction:column;gap:0.375rem;margin-bottom:1rem; }
.label         { font-size:0.875rem;font-weight:500; }
.btn-row       { display:flex;gap:0.75rem;margin-top:1.5rem; }

/* ── Action buttons ────────────────────────────────── */
.action-row { display:flex;gap:0.5rem;align-items:center; }
.btn-xs     { padding:0.25rem 0.625rem;border-radius:0.375rem;font-size:0.75rem;font-weight:600;cursor:pointer;border:1px solid var(--color-border);background:var(--color-bg);transition:opacity 0.15s; }
.btn-xs:hover { opacity: 0.8; }
.btn-green  { background:#d1fae5;color:#065f46;border-color:#6ee7b7; }
.btn-red    { background:#fee2e2;color:#991b1b;border-color:#fca5a5; }
.btn-sm     { padding:0.375rem 0.875rem;font-size:0.8rem; }

/* ── Pagination ────────────────────────────────────── */
.pagination { display:flex;align-items:center;gap:1rem;justify-content:center;padding:1.5rem 0; }

/* ── Alerts ────────────────────────────────────────── */
.alert       { padding:0.75rem 1rem;border-radius:0.5rem;margin-bottom:1rem;font-size:0.875rem; }
.alert-green { background:#d1fae5;color:#065f46;border:1px solid #6ee7b7; }
.alert-red   { background:#fee2e2;color:#991b1b;border:1px solid #fca5a5; }

/* ── Quick links ───────────────────────────────────── */
.quick-links { display:flex;gap:0.75rem;flex-wrap:wrap;margin:1.5rem 0; }

/* ── Log section ───────────────────────────────────── */
.log-header { display:flex;justify-content:space-between;align-items:center;margin-bottom:0.75rem; }

/* ── Typography helpers ────────────────────────────── */
.text-muted    { color: var(--color-muted-fg); }
.text-sm       { font-size: 0.875rem; }
.font-medium   { font-weight: 500; }
```

---

## 13. Implementation Order

| # | Task | Files | Est. |
|---|------|-------|------|
| 1 | Role rename + middleware | `migrations/005_rename_superadmin.sql`, `middleware/auth.go` | 30m |
| 2 | Audit log | `migrations/006_audit_log.sql`, `models/audit.go`, `handlers/audit.go`, `views/audit_log.templ` | 2h |
| 3 | Dashboard stats | `models/dashboard.go`, `handlers/dashboard.go`, `views/admin_dashboard.templ` | 1h |
| 4 | Tenant list | `models/tenant_list.go`, update `handlers/tenants.go`, `views/tenant_list.templ` | 2h |
| 5 | Health check | `handlers/health.go`, update `show_tenant.templ` | 1h |
| 6 | SSE live logs | `handlers/logs.go`, `views/log_viewer.templ` | 2h |
| 7 | Settings | `migrations/007_settings.sql`, `models/settings.go`, `handlers/settings.go`, `views/smtp_settings.templ`, `views/provisioner_settings.templ` | 3h |
| 8 | Impersonation | `handlers/impersonate.go`, update `client_layout.templ`, update `show_tenant.templ` | 2h |
| 9 | CSV export | `handlers/export.go` | 30m |
| 10 | Shared components | `views/shared.templ` | 30m |
| 11 | Wire up routes + handler struct | `main.go`, `handlers/handler.go` | 30m |
| 12 | CSS | `static/css/admin.css` | 30m |
| 13 | Git cleanup | `.gitignore` | 10m |

**Total: ~16h — do sections 1 → 4 first, they unblock everything else.**

---

> **Golden rule:** every handler that changes state must call `h.audit.Log(...)`.  
> Run migrations after each step: `go run ./cmd/migrate up`  
> Test each endpoint with `curl` or the browser before moving to the next.
