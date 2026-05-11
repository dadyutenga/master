Task 7 - Audit Log (Admin Actions)

Objective:
Add a persistent audit log to track admin actions with tenant context, IP address, and timestamps.

Stack: Go + Fiber + templ + SQLite

DO NOT block primary actions if audit logging fails.

---

### 1) Migration

Create migrations/006_audit_log.sql:
```sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id   INTEGER NOT NULL,
    action     TEXT    NOT NULL,
    tenant_id  INTEGER,
    detail     TEXT    DEFAULT '',
    ip_address TEXT    DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES users(id)
);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_tenant_id  ON audit_logs(tenant_id);
```

Run migrations:
```sh
go run ./cmd/migrate up
```

---

### 2) Model

Create internal/models/audit.go:
```go
package models

import "time"

type AuditLog struct {
    ID         int64
    AdminID    int64
    AdminEmail string
    Action     string
    TenantID   *int64
    Detail     string
    IPAddress  string
    CreatedAt  time.Time
}
```

Note: TenantID is a pointer so NULL scans cleanly in SQLite.

---

### 3) Helper

Create internal/handlers/audit.go to avoid circular deps:
```go
package handlers

import (
    "database/sql"
    "log"
)

// LogAction inserts one audit row. It intentionally swallows errors.
func LogAction(db *sql.DB, adminID int64, action string, tenantID *int64, detail, ip string) {
    _, err := db.Exec(
        `INSERT INTO audit_logs (admin_id, action, tenant_id, detail, ip_address)
         VALUES (?, ?, ?, ?, ?)`,
        adminID, action, tenantID, detail, ip,
    )
    if err != nil {
        log.Printf("audit log error: %v", err)
    }
}
```

---

### 4) Routes

Add to main.go inside the admin group:
```go
admin.Get("/audit",        h.AuditLog)
admin.Get("/audit/export", h.ExportAuditCSV)
```

---

### 5) AuditLog handler (list + pagination)

In internal/handlers/admin.go:
```go
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
    return render(c, views.AuditLogPage(logs, page))
}
```

Imports needed:
- strconv
- github.com/dadyutenga/hms-control/internal/models

---

### 6) Where to call LogAction

Call LogAction after successful DB writes in these handlers:
- ApproveTenant: action "tenant.approved"
- SuspendTenant: action "tenant.suspended"
- RetryProvision: action "provision.retry"
- StartTenantDeployment: action "deployment.start"
- StopTenantDeployment: action "deployment.stop"
- UpdateTenantBilling: action "billing.updated"
- UpdateContactSettings: action "settings.contact"

Pattern:
```go
sess, _ := h.store.Get(c)
adminID := sess.Get("user_id").(int64)
LogAction(h.db, adminID, "tenant.approved", &tenantIDInt, "", c.IP())
```

---

### 7) Audit CSV export

Implement ExportAuditCSV in internal/handlers/admin.go:
```go
func (h *Handler) ExportAuditCSV(c *fiber.Ctx) error {
    c.Set("Content-Type",        "text/csv")
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
            tenantID                           *int64
        )
        if err := rows.Scan(&id, &adminEmail, &action, &tenantID, &detail, &ip, &ts); err != nil {
            return err
        }
        tid := ""
        if tenantID != nil {
            tid = strconv.FormatInt(*tenantID, 10)
        }
        w.Write([]string{
            strconv.FormatInt(id, 10), adminEmail, action, tid, detail, ip, ts,
        })
    }

    w.Flush()
    return w.Error()
}
```

Imports needed:
- encoding/csv
- strconv

---

### 8) Acceptance Criteria

1) Audit logs created for all listed actions.
2) Logs visible in admin UI.
3) CSV export downloads correctly.
4) Audit errors do not break primary operations.

Expected Outcome:
- Admin actions are fully auditable.

Priority:
HIGH
