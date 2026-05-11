Task 16 - Final Route Block, Imports, Migration Order, Pitfalls

Objective:
Validate the final admin route block, imports, migration order, and common pitfalls.

---

### 1) Final admin route group (main.go)

After all steps, admin routes should be:
```go
admin := app.Group("/admin",
    middleware.Auth(store, database),
    middleware.RequireRole("admin"),
)
admin.Get("/",  h.AdminDashboard)
admin.Get("/audit",        h.AuditLog)
admin.Get("/audit/export", h.ExportAuditCSV)
// /tenants/export MUST come before /tenants/:id
admin.Get("/tenants/export",             h.ExportTenantsCSV)
admin.Get("/tenants",                    h.ListTenants)
admin.Get("/tenants/:id",                h.ShowTenant)
admin.Get("/tenants/:id/deployments/:deploymentId", h.ShowDeployment)
admin.Get("/tenants/:id/health",         h.TenantHealthCheck)
admin.Get("/tenants/:id/logs/stream",    h.StreamProvisionLogs)
admin.Post("/tenants/:id/approve",          h.ApproveTenant)
admin.Post("/tenants/:id/suspend",          h.SuspendTenant)
admin.Post("/tenants/:id/retry",            h.RetryProvision)
admin.Post("/tenants/:id/deployments/start",h.StartTenantDeployment)
admin.Post("/tenants/:id/deployments/stop", h.StopTenantDeployment)
admin.Post("/tenants/:id/billing",          h.UpdateTenantBilling)
admin.Post("/tenants/:id/impersonate",      h.ImpersonateTenant)
admin.Post("/impersonate/stop",             h.StopImpersonation)
admin.Get("/settings/contact",         h.AdminContactSettings)
admin.Post("/settings/contact",        h.UpdateContactSettings)
admin.Get("/settings/smtp",            h.AdminSMTPSettings)
admin.Post("/settings/smtp",           h.UpdateSMTPSettings)
admin.Post("/settings/smtp/test",      h.TestSMTP)
admin.Get("/settings/provisioner",     h.AdminProvisionerSettings)
admin.Post("/settings/provisioner",    h.UpdateProvisionerSettings)
```

---

### 2) Admin handler imports (internal/handlers/admin.go)

```go
import (
    "bufio"
    "database/sql"
    "encoding/csv"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/dadyutenga/hms-control/internal/models"
    views "github.com/dadyutenga/hms-control/internal/views/admin"
)
```

---

### 3) Migration order

Run:
```sh
go run ./cmd/migrate up
```

Files run in order:
- 005_rename_superadmin.sql (role -> admin)
- 006_audit_log.sql (audit_logs table + indexes)
- 007_settings.sql (settings table + defaults)

---

### 4) Common pitfalls

1) Route order: /tenants/export must be registered before /tenants/:id.
2) Session type assertion: guard sess.Get("user_id").(int64) to avoid panic.
3) NULL scanning: use pointer types for nullable SQLite columns.
4) SSE behind nginx: X-Accel-Buffering: no is required.
5) SMTP password: UpdateSMTPSettings must not clear password when blank.

Expected Outcome:
- Final admin setup is stable and consistent.

Priority:
LOW
