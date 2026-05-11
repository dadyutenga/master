

HMS Control — Admin Extension: Full Implementation
## Guide
Stack: Go + Fiber + templ + SQLite
Module: github.com/dadyutenga/hms-control
Run: go run ./cmd/migrate up then air
## Pre-flight Checklist
Before starting, confirm these files exist and note their exact paths — every code snippet
below assumes this layout:
main.go
internal/
config/
db/
handlers/      ← admin.go lives here
middleware/    ← auth.go lives here
models/        ← one file per model
mailer/
provisioner/
migrations/      ← numbered .sql files
Step 1 — Role Rename (superadmin → admin)
Time: 30 min
1a. internal/middleware/auth.go
Find every call to RequireRole and change the string literal:
## // BEFORE
middleware.RequireRole("superadmin")
## // AFTER
middleware.RequireRole("admin")

Also update the role check inside RequireRole itself if it does a string comparison:
func RequireRole(role string) fiber.Handler {
return func(c *fiber.Ctx) error {
sess, err := store.Get(c)
if err != nil {
return c.Redirect("/login")
## }
userRole, ok := sess.Get("role").(string)
if !ok || userRole != role {
return c.Status(fiber.StatusForbidden).SendString("Forbidden")
## }
return c.Next()
## }
## }
1b. main.go
The existing admin group in main.go (line 74) uses RequireRole("superadmin"). Update it:
admin := app.Group("/admin",
middleware.Auth(store, database),
middleware.RequireRole("admin"),   // ← was "superadmin"
## )
1c. Migration — migrations/005_rename_superadmin.sql
Create this file. The migrate runner picks files in numeric order:
## -- 005_rename_superadmin.sql
UPDATE users SET role = 'admin' WHERE role = 'superadmin';
Run: go run ./cmd/migrate up
Verify: Log in with an existing superadmin account — the dashboard should still be
accessible with no 403.
## Step 2 — Audit Log
Time: 2 hrs

2a. Migration — migrations/006_audit_log.sql
## -- 006_audit_log.sql
CREATE TABLE IF NOT EXISTS audit_logs (
id         INTEGER PRIMARY KEY AUTOINCREMENT,
admin_id   INTEGER NOT NULL,
action     TEXT    NOT NULL,
tenant_id  INTEGER,
detail     TEXT    DEFAULT '',
ip_address TEXT    DEFAULT '',
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (admin_id) REFERENCES users(id)
## );
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_tenant_id  ON audit_logs(tenant_id);
2b. Model — internal/models/audit.go
package models
import "time"
type AuditLog struct {
ID        int64
AdminID   int64
AdminEmail string  // populated by JOIN in queries
Action    string
TenantID  *int64  // pointer — can be NULL
Detail    string
IPAddress string
CreatedAt time.Time
## }
Why pointer for TenantID? SQLite NULL cannot scan into int64. Using *int64 means
a NULL row scans to nil with no error.
2c. Helper — internal/handlers/audit.go
Create a separate file for the helper so it is importable from every handler file without
circular deps:
package handlers

2d. Routes — add to main.go inside the admin group
admin.Get("/audit",        h.AuditLog)
admin.Get("/audit/export", h.ExportAuditCSV)
Place these before log.Fatal(app.Listen(":8080")).
2e. Handler — AuditLog in internal/handlers/admin.go
import (
## "database/sql"
## "log"
## )
// LogAction inserts one audit row. It intentionally swallows errors — a
// failed audit log must never break the primary operation.
func LogAction(db *sql.DB, adminID int64, action string, tenantID *int64, detail, ip string) {
_, err := db.Exec(
`INSERT INTO audit_logs (admin_id, action, tenant_id, detail, ip_address)
## VALUES (?, ?, ?, ?, ?)`,
adminID, action, tenantID, detail, ip,
## )
if err != nil {
log.Printf("audit log error: %v", err)
## }
## }
func (h *Handler) AuditLog(c *fiber.Ctx) error {
page, _ := strconv.Atoi(c.Query("page", "1"))
if page < 1 {
page = 1
## }
const limit = 50
offset := (page - 1) * limit
rows, err := h.db.Query(`
SELECT a.id, u.email, a.action, a.tenant_id, a.detail, a.ip_address, a.created_at
FROM audit_logs a
JOIN users u ON u.id = a.admin_id
ORDER BY a.created_at DESC
## LIMIT ? OFFSET ?
`, limit, offset)
if err != nil {
return err
## }

Import note: add "strconv" and "github.com/dadyutenga/hms-
control/internal/models" to the import block.
2f. Where to call LogAction — inside existing handlers
For each handler below, call LogAction after the DB write succeeds. Pull adminID from the
session and use c.IP() for the IP.
Pattern (repeat for each handler):
func (h *Handler) ApproveTenant(c *fiber.Ctx) error {
sess, _ := h.store.Get(c)
adminID := sess.Get("user_id").(int64)
id := c.Params("id")
tenantIDInt, _ := strconv.ParseInt(id, 10, 64)
// ... existing approval logic ...
LogAction(h.db, adminID, "tenant.approved", &tenantIDInt, "", c.IP())
return c.Redirect("/admin/tenants/" + id)
## }
HandlerAction string
defer rows.Close()
var logs []models.AuditLog
for rows.Next() {
var l models.AuditLog
if err := rows.Scan(
&l.ID, &l.AdminEmail, &l.Action,
&l.TenantID, &l.Detail, &l.IPAddress, &l.CreatedAt,
); err != nil {
return err
## }
logs = append(logs, l)
## }
if err := rows.Err(); err != nil {
return err
## }
return render(c, views.AuditLogPage(logs, page))
## }

ApproveTenant"tenant.approved"
SuspendTenant"tenant.suspended"
RetryProvision"provision.retry"
StartTenantDeployment"deployment.start"
StopTenantDeployment"deployment.stop"
UpdateTenantBilling"billing.updated"
UpdateContactSettings"settings.contact"
## Step 3 — Dashboard Stats
Time: 1 hr
3a. Model — internal/models/dashboard.go
package models
type DashboardStats struct {
TotalTenants       int
ActiveTenants      int
PendingTenants     int
SuspendedTenants   int
RunningDeployments int
FailedDeployments  int
RecentActions      []AuditLog
## }
3b. Handler — update AdminDashboard in internal/handlers/admin.go
func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
var stats models.DashboardStats
// Tenant counts — single pass with CASE SUM
err := h.db.QueryRow(`
## SELECT
## COUNT(*),
COALESCE(SUM(CASE WHEN status='active'    THEN 1 ELSE 0 END), 0),

COALESCE(SUM(CASE WHEN status='pending'   THEN 1 ELSE 0 END), 0),
COALESCE(SUM(CASE WHEN status='suspended' THEN 1 ELSE 0 END), 0)
FROM tenants
`).Scan(
&stats.TotalTenants,
&stats.ActiveTenants,
&stats.PendingTenants,
&stats.SuspendedTenants,
## )
if err != nil {
return err
## }
// Deployment counts
err = h.db.QueryRow(`
## SELECT
COALESCE(SUM(CASE WHEN status='running' THEN 1 ELSE 0 END), 0),
COALESCE(SUM(CASE WHEN status='failed'  THEN 1 ELSE 0 END), 0)
FROM deployments
`).Scan(
&stats.RunningDeployments,
&stats.FailedDeployments,
## )
if err != nil {
return err
## }
// Last 10 audit actions
rows, err := h.db.Query(`
SELECT a.id, u.email, a.action, a.tenant_id, a.detail, a.ip_address, a.created_at
FROM audit_logs a
JOIN users u ON u.id = a.admin_id
ORDER BY a.created_at DESC
## LIMIT 10
## `)
if err != nil {
return err
## }
defer rows.Close()
for rows.Next() {
var l models.AuditLog
if err := rows.Scan(
&l.ID, &l.AdminEmail, &l.Action,
&l.TenantID, &l.Detail, &l.IPAddress, &l.CreatedAt,
); err != nil {
return err

COALESCE prevents NULL panics when tables are empty.
## Step 4 — Tenant Search, Filter, Pagination
Time: 2 hrs
4a. Handler — update ListTenants in internal/handlers/admin.go
## }
stats.RecentActions = append(stats.RecentActions, l)
## }
return render(c, views.AdminDashboardPage(stats))
## }
func (h *Handler) ListTenants(c *fiber.Ctx) error {
status := c.Query("status", "")
search := c.Query("q", "")
page, _ := strconv.Atoi(c.Query("page", "1"))
if page < 1 {
page = 1
## }
const limit = 20
offset := (page - 1) * limit
// Build query dynamically — safe because we only append ? placeholders
query := `SELECT id, name, email, status, created_at FROM tenants WHERE 1=1`
args := []interface{}{}
if status != "" {
query += " AND status = ?"
args = append(args, status)
## }
if search != "" {
query += " AND (name LIKE ? OR email LIKE ?)"
args = append(args, "%"+search+"%", "%"+search+"%")
## }
// Count total rows for pagination (same filters, no LIMIT)
countQuery := "SELECT COUNT(*) FROM tenants WHERE 1=1"
if status != "" {
countQuery += " AND status = ?"
## }
if search != "" {

## Step 5 — Deployment Health Check
Time: 1 hr
5a. Route — add to main.go
admin.Get("/tenants/:id/health", h.TenantHealthCheck)
5b. Handler — TenantHealthCheck in internal/handlers/admin.go
countQuery += " AND (name LIKE ? OR email LIKE ?)"
## }
var total int
h.db.QueryRow(countQuery, args...).Scan(&total)
query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
args = append(args, limit, offset)
rows, err := h.db.Query(query, args...)
if err != nil {
return err
## }
defer rows.Close()
var tenants []models.Tenant
for rows.Next() {
var t models.Tenant
if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status, &t.CreatedAt); err != nil {
return err
## }
tenants = append(tenants, t)
## }
totalPages := (total + limit - 1) / limit
return render(c, views.ListTenantsPage(tenants, search, status, page, totalPages))
## }
func (h *Handler) TenantHealthCheck(c *fiber.Ctx) error {
tenantID := c.Params("id")
var endpoint string

Import note: add "net/http" and "time" to imports.
5c. Templ template snippet — ShowTenant page
In the tenant detail templ file, add the health badge and the fetch call:
err := h.db.QueryRow(
`SELECT endpoint FROM deployments WHERE tenant_id = ? AND status = 'running' LIMIT 1`,
tenantID,
).Scan(&endpoint)
// No running deployment — return DOWN immediately, not an error
if err != nil {
return c.JSON(fiber.Map{"status": "DOWN", "tenant_id": tenantID, "reason": "no running deployment"})
## }
client := &http.Client{Timeout: 5 * time.Second}
resp, err := client.Get(endpoint + "/health")
status := "DOWN"
if err == nil && resp.StatusCode == http.StatusOK {
status = "UP"
resp.Body.Close()
## }
return c.JSON(fiber.Map{"status": status, "tenant_id": tenantID})
## }
<span id="health-badge" class="badge badge-grey">checking...</span>
## <script>
fetch('/admin/tenants/{{ .Tenant.ID }}/health')
.then(r => r.json())
## .then(d => {
const badge = document.getElementById('health-badge');
badge.textContent = d.status;
badge.className = d.status === 'UP' ? 'badge badge-green' : 'badge badge-red';
## })
## .catch(() => {
document.getElementById('health-badge').textContent = 'ERROR';
## });
## </script>

Step 6 — SSE Live Provisioning Logs
Time: 2 hrs
6a. Route — add to main.go
admin.Get("/tenants/:id/logs/stream", h.StreamProvisionLogs)
6b. Handler — StreamProvisionLogs in internal/handlers/admin.go
func (h *Handler) StreamProvisionLogs(c *fiber.Ctx) error {
tenantID := c.Params("id")
logPath  := fmt.Sprintf("./tmp/provision-%s.log", tenantID)
c.Set("Content-Type",  "text/event-stream")
c.Set("Cache-Control", "no-cache")
c.Set("Connection",    "keep-alive")
c.Set("X-Accel-Buffering", "no") // disable nginx buffering if behind a proxy
c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
f, err := os.Open(logPath)
if err != nil {
fmt.Fprintf(w, "data: log file not found for tenant %s\n\n", tenantID)
w.Flush()
return
## }
defer f.Close()
reader := bufio.NewReader(f)
for {
// Check if client disconnected
select {
case <-c.Context().Done():
return
default:
## }
line, err := reader.ReadString('\n')
if len(line) > 0 {
fmt.Fprintf(w, "data: %s\n\n", strings.TrimRight(line, "\r\n"))
if flushErr := w.Flush(); flushErr != nil {
return // client gone
## }
## }
if err != nil {

Import note: add "bufio", "fmt", "os", "strings", "time" to imports.
6c. Templ snippet — deployment detail page
Step 7 — Settings Expansion (SMTP + Provisioner)
Time: 3 hrs
7a. Migration — migrations/007_settings.sql
## -- 007_settings.sql
CREATE TABLE IF NOT EXISTS settings (
// EOF — wait for more content
time.Sleep(500 * time.Millisecond)
## }
## }
## })
return nil
## }
<pre id="log-output"
style="height:300px;overflow-y:auto;background:#1e1e1e;color:#d4d4d4;padding:1rem;font-family:monospace;">
## </pre>
## <script>
## (function() {
const out = document.getElementById('log-output');
const es = new EventSource('/admin/tenants/{{ .TenantID }}/logs/stream');
es.onmessage = function(e) {
out.textContent += e.data + '\n';
out.scrollTop = out.scrollHeight;
## };
es.onerror = function() {
out.textContent += '\n[stream closed]\n';
es.close();
## };
## })();
## </script>

key        TEXT PRIMARY KEY,
value      TEXT NOT NULL DEFAULT '',
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
## );
INSERT OR IGNORE INTO settings (key, value) VALUES
## ('smtp_host',        ''),
## ('smtp_port',        '587'),
## ('smtp_user',        ''),
## ('smtp_pass',        ''),
## ('smtp_from',        'noreply@localhost'),
## ('provision_script', './scripts/provision.sh'),
## ('docker_template',  'default');
7b. Routes — add to main.go
admin.Get("/settings/smtp",         h.AdminSMTPSettings)
admin.Post("/settings/smtp",        h.UpdateSMTPSettings)
admin.Post("/settings/smtp/test",   h.TestSMTP)
admin.Get("/settings/provisioner",  h.AdminProvisionerSettings)
admin.Post("/settings/provisioner", h.UpdateProvisionerSettings)
7c. Handler helpers — reading/writing settings
Add these two private helpers to internal/handlers/admin.go. Every settings handler uses
them:
// getSetting reads one key. Returns "" on miss — never errors to caller.
func (h *Handler) getSetting(key string) string {
var value string
h.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
return value
## }
// setSetting upserts one key.
func (h *Handler) setSetting(key, value string) error {
_, err := h.db.Exec(`
INSERT INTO settings (key, value, updated_at)
## VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, key, value)
return err
## }

7d. Handler — AdminSMTPSettings
func (h *Handler) AdminSMTPSettings(c *fiber.Ctx) error {
settings := map[string]string{
"smtp_host": h.getSetting("smtp_host"),
"smtp_port": h.getSetting("smtp_port"),
"smtp_user": h.getSetting("smtp_user"),
"smtp_from": h.getSetting("smtp_from"),
## }
return render(c, views.SMTPSettingsPage(settings))
## }
7e. Handler — UpdateSMTPSettings
func (h *Handler) UpdateSMTPSettings(c *fiber.Ctx) error {
fields := map[string]string{
"smtp_host": c.FormValue("smtp_host"),
"smtp_port": c.FormValue("smtp_port"),
"smtp_user": c.FormValue("smtp_user"),
"smtp_from": c.FormValue("smtp_from"),
## }
// Only update password if provided (avoid clearing it accidentally)
if pass := c.FormValue("smtp_pass"); pass != "" {
fields["smtp_pass"] = pass
## }
for key, value := range fields {
if err := h.setSetting(key, value); err != nil {
return err
## }
## }
// Log the settings change
sess, _ := h.store.Get(c)
adminID := sess.Get("user_id").(int64)
LogAction(h.db, adminID, "settings.smtp_updated", nil, "", c.IP())
return c.Redirect("/admin/settings/smtp")
## }
7f. Handler — TestSMTP
func (h *Handler) TestSMTP(c *fiber.Ctx) error {
to := c.FormValue("test_email")

7g. Handler — AdminProvisionerSettings and UpdateProvisionerSettings
func (h *Handler) AdminProvisionerSettings(c *fiber.Ctx) error {
settings := map[string]string{
"provision_script": h.getSetting("provision_script"),
"docker_template":  h.getSetting("docker_template"),
## }
return render(c, views.ProvisionerSettingsPage(settings))
## }
func (h *Handler) UpdateProvisionerSettings(c *fiber.Ctx) error {
fields := map[string]string{
"provision_script": c.FormValue("provision_script"),
"docker_template":  c.FormValue("docker_template"),
## }
for key, value := range fields {
if err := h.setSetting(key, value); err != nil {
return err
## }
## }
sess, _ := h.store.Get(c)
adminID := sess.Get("user_id").(int64)
LogAction(h.db, adminID, "settings.provisioner_updated", nil, "", c.IP())
return c.Redirect("/admin/settings/provisioner")
## }
## Step 8 — Tenant Impersonation
Time: 2 hrs
if to == "" {
return c.JSON(fiber.Map{"ok": false, "error": "test_email is required"})
## }
err := h.mail.Send(to, "HMS SMTP Test", "If you see this, SMTP is configured correctly.")
if err != nil {
return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
## }
return c.JSON(fiber.Map{"ok": true})
## }

8a. Routes — add to main.go
admin.Post("/tenants/:id/impersonate", h.ImpersonateTenant)
admin.Post("/impersonate/stop",        h.StopImpersonation)
## 8b. Handlers
func (h *Handler) ImpersonateTenant(c *fiber.Ctx) error {
sess, err := h.store.Get(c)
if err != nil {
return c.Redirect("/login")
## }
tenantIDStr := c.Params("id")
tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
if err != nil {
return fiber.ErrBadRequest
## }
// Confirm the tenant actually exists before impersonating
var exists int
h.db.QueryRow(`SELECT COUNT(*) FROM tenants WHERE id = ?`, tenantID).Scan(&exists)
if exists == 0 {
return fiber.ErrNotFound
## }
adminID := sess.Get("user_id").(int64)
// Stash the real admin ID so StopImpersonation can restore it
sess.Set("original_admin_id",      adminID)
sess.Set("impersonating_tenant_id", tenantID)
LogAction(h.db, adminID, "tenant.impersonated", &tenantID, "", c.IP())
if err := sess.Save(); err != nil {
return err
## }
return c.Redirect("/dashboard")
## }
func (h *Handler) StopImpersonation(c *fiber.Ctx) error {
sess, err := h.store.Get(c)
if err != nil {
return c.Redirect("/login")
## }

8c. Middleware — internal/middleware/auth.go
Update the Auth middleware to swap identity when impersonating. The key logic goes
after the normal user lookup:
Note: adapt the tenant user lookup to match your actual users table schema. If users
are linked to tenants differently, adjust the query accordingly.
sess.Delete("impersonating_tenant_id")
// Restore the admin's original user_id in session
if origID := sess.Get("original_admin_id"); origID != nil {
sess.Set("user_id", origID)
sess.Delete("original_admin_id")
## }
if err := sess.Save(); err != nil {
return err
## }
return c.Redirect("/admin/tenants")
## }
func Auth(store *session.Store, db *sql.DB) fiber.Handler {
return func(c *fiber.Ctx) error {
sess, err := store.Get(c)
if err != nil || sess.Get("user_id") == nil {
return c.Redirect("/login")
## }
userID := sess.Get("user_id").(int64)
// Impersonation: if set, load the tenant's primary user instead
if impersonatedTenantID := sess.Get("impersonating_tenant_id"); impersonatedTenantID != nil {
var tenantUserID int64
err := db.QueryRow(
`SELECT id FROM users WHERE tenant_id = ? LIMIT 1`,
impersonatedTenantID,
).Scan(&tenantUserID)
if err == nil {
userID = tenantUserID
## }
## }
c.Locals("user_id", userID)
return c.Next()
## }
## }

8d. Visible indicator (templ)
Add a banner to your base layout so the admin always knows they are impersonating:
Step 9 — Export Tenants CSV
Time: 1 hr
9a. Route — add to main.go
// IMPORTANT: place /tenants/export BEFORE /tenants/:id
// Fiber matches routes in registration order — /tenants/export would
// otherwise be caught by /tenants/:id with id="export".
admin.Get("/tenants/export", h.ExportTenantsCSV)
admin.Get("/tenants/:id",    h.ShowTenant)
9b. Handler — ExportTenantsCSV in internal/handlers/admin.go
{{ if .ImpersonatingTenantID }}
<div style="background:#f59e0b;color:#000;padding:0.5rem 1rem;text-align:center;">
⚠ You are impersonating tenant #{{ .ImpersonatingTenantID }}
<form method="POST" action="/impersonate/stop" style="display:inline;margin-left:1rem;">
<button type="submit">Stop Impersonation</button>
## </form>
## </div>
{{ end }}
func (h *Handler) ExportTenantsCSV(c *fiber.Ctx) error {
c.Set("Content-Type",        "text/csv")
c.Set("Content-Disposition", `attachment; filename="tenants.csv"`)
rows, err := h.db.Query(
`SELECT id, name, email, status, created_at FROM tenants ORDER BY created_at DESC`,
## )
if err != nil {
return err
## }
defer rows.Close()
w := csv.NewWriter(c.Response().BodyWriter())
// Header row
if err := w.Write([]string{"ID", "Name", "Email", "Status", "Created At"}); err != nil {

Import note: add "encoding/csv" to imports.
9c. Also add the audit CSV export handler
The plan’s route GET /audit/export was registered in Step 2d. Here is the handler:
return err
## }
for rows.Next() {
var (
id        int64
name, email, status, createdAt string
## )
if err := rows.Scan(&id, &name, &email, &status, &createdAt); err != nil {
return err
## }
if err := w.Write([]string{
strconv.FormatInt(id, 10), name, email, status, createdAt,
}); err != nil {
return err
## }
## }
w.Flush()
return w.Error() // returns nil on success
## }
func (h *Handler) ExportAuditCSV(c *fiber.Ctx) error {
c.Set("Content-Type",        "text/csv")
c.Set("Content-Disposition", `attachment; filename="audit_log.csv"`)
rows, err := h.db.Query(`
SELECT a.id, u.email, a.action, a.tenant_id, a.detail, a.ip_address, a.created_at
FROM audit_logs a
JOIN users u ON u.id = a.admin_id
ORDER BY a.created_at DESC
## `)
if err != nil {
return err
## }
defer rows.Close()
w := csv.NewWriter(c.Response().BodyWriter())
w.Write([]string{"ID", "Admin Email", "Action", "Tenant ID", "Detail", "IP", "Created At"})

## Step 10 — Git Cleanup
Time: 10 min
# Check what is currently tracked that shouldn't be
git ls-files | grep -E "^(hms-control$|hms-control\.exe|hms\.log|tmp/)"
# Add to .gitignore (idempotent — won't duplicate if already there)
grep -qxF 'hms-control'     .gitignore || echo 'hms-control'     >> .gitignore
grep -qxF 'hms-control.exe' .gitignore || echo 'hms-control.exe' >> .gitignore
grep -qxF 'hms.log'         .gitignore || echo 'hms.log'         >> .gitignore
grep -qxF 'tmp/'            .gitignore || echo 'tmp/'            >> .gitignore
# Remove from tracking (keeps files on disk)
git rm --cached hms-control hms-control.exe hms.log 2>/dev/null || true
git rm -r --cached tmp/ 2>/dev/null || true
git add .gitignore
git commit -m "chore: remove binaries and logs from tracking"
for rows.Next() {
var (
id                                    int64
adminEmail, action, detail, ip, ts    string
tenantID                              *int64
## )
if err := rows.Scan(&id, &adminEmail, &action, &tenantID, &detail, &ip, &ts); err != nil {
return err
## }
tid := ""
if tenantID != nil {
tid = strconv.FormatInt(*tenantID, 10)
## }
w.Write([]string{
strconv.FormatInt(id, 10), adminEmail, action, tid, detail, ip, ts,
## })
## }
w.Flush()
return w.Error()
## }

Final main.go — complete admin route block
After all steps, your admin route group in main.go should look exactly like this:
admin := app.Group("/admin",
middleware.Auth(store, database),
middleware.RequireRole("admin"),
## )
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
Full Import List for internal/handlers/admin.go
import (
## "bufio"

## "database/sql"
## "encoding/csv"
## "fmt"
## "log"
## "net/http"
## "os"
## "strconv"
## "strings"
## "time"
## "github.com/gofiber/fiber/v2"
## "github.com/dadyutenga/hms-control/internal/models"
views "github.com/dadyutenga/hms-control/internal/views/admin"
## )
## Migration Run Order
go run ./cmd/migrate up
Files run in this order:
FilePurpose
## 005_rename_superadmin.sql
role → admin
## 006_audit_log.sql
audit_logs table + indexes
## 007_settings.sql
settings table + defaults
## Common Pitfalls
Fiber route order matters. /tenants/export must be registered before /tenants/:id or
Fiber will match export as the :id param and call ShowTenant instead.
Session type assertion panics. sess.Get("user_id").(int64) will panic if the key is
missing. Guard it:
uid, ok := sess.Get("user_id").(int64)
if !ok {
return c.Redirect("/login")

## }
NULL scanning. Any column that can be NULL in SQLite must scan into a pointer (*int64,
*string) or sql.NullString / sql.NullInt64. The tenant_id in audit_logs is nullable
— the model uses *int64 for this reason.
SSE and nginx. If you deploy behind nginx, you must set X-Accel-Buffering: no (already
included in the handler above) or nginx buffers the stream and the browser sees nothing
until the connection closes.
SMTP password. The UpdateSMTPSettings handler skips the password update if the form
field is blank — this prevents wiping the stored password on every save when the field is left
empty.