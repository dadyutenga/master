# HMS Master — Go SaaS Platform
### Complete Architecture & Implementation Guide (Go + Fiber + Templ + HTMX)

> **Author:** Dady Utenga · BIG LITE CODE · Iringa, Tanzania
> **Control Plane:** Go + Fiber + Templ + HTMX + PostgreSQL
> **Tenant:** Laravel (HMS monolith) — untouched
> **Infra:** On-premise · Docker Compose per tenant · Traefik
> **Trigger:** Email verification → Superadmin approval → Go provisioner
> **Version:** 2.0.0

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Go Microservice Project Structure](#2-go-microservice-project-structure)
3. [Database Design](#3-database-design)
4. [Go Models & DB Layer (sqlc)](#4-go-models--db-layer-sqlc)
5. [Fiber App Bootstrap](#5-fiber-app-bootstrap)
6. [Auth — Registration & Email Verification](#6-auth--registration--email-verification)
7. [Provisioner — Core Engine](#7-provisioner--core-engine)
8. [Provisioner Worker (Goroutine Pool)](#8-provisioner-worker-goroutine-pool)
9. [Superadmin Handlers](#9-superadmin-handlers)
10. [Templ + HTMX Views](#10-templ--htmx-views)
11. [Email Service](#11-email-service)
12. [Docker Templates](#12-docker-templates)
13. [Provision Shell Script](#13-provision-shell-script)
14. [Traefik & DNS](#14-traefik--dns)
15. [Deployment & Server Setup](#15-deployment--server-setup)
16. [Updating All Tenants](#16-updating-all-tenants)

---

## 1. System Overview

### Why Go for the Control Plane

| Concern | Laravel Master | Go Microservice |
|---|---|---|
| Startup time | ~800ms | ~5ms |
| Provisioning goroutines | PHP queue worker process | Native goroutines, no extra process |
| Binary deployment | Composer + PHP runtime | Single compiled binary |
| Concurrency | Queue jobs, one at a time per worker | Goroutine pool, N parallel provisions |
| Shell exec | `exec()` | `os/exec` with streaming stdout |
| Memory | ~80MB idle | ~12MB idle |

### Full Architecture

```
Internet
    │
Traefik (80/443) — wildcard SSL via Let's Encrypt
    │
    ├── master.hms.co.tz   →  Go/Fiber Control Plane  (this service)
    │                          ├── Auth (register, verify, login)
    │                          ├── Superadmin dashboard (Templ+HTMX)
    │                          ├── Client portal
    │                          └── Provisioner engine (goroutine pool)
    │
    ├── dady2.hms.co.tz    →  Laravel HMS Container A  →  PostgreSQL A
    ├── dady3.hms.co.tz    →  Laravel HMS Container B  →  PostgreSQL B
    └── clientN.hms.co.tz  →  Laravel HMS Container N  →  PostgreSQL N

On-premise server
├── /opt/hms-control/          ← Go binary + templates + scripts
│   ├── hms-control            ← compiled binary
│   ├── scripts/
│   │   ├── provision.sh
│   │   └── update_all.sh
│   └── docker-templates/
│       ├── docker-compose.template.yml
│       └── .env.template
├── /opt/hms-source/           ← Laravel HMS codebase (read-only)
├── /opt/tenants/              ← per-tenant docker-compose + .env + volumes
│   ├── dady2/
│   └── dady3/
└── /opt/traefik/
    └── docker-compose.yml
```

### Provisioning Flow

```
POST /register
    → create user + tenant record (status: pending_verification)
    → send verification email (Go mail goroutine)

GET /verify/:token
    → mark email verified
    → update tenant status: pending_approval

Superadmin: POST /admin/tenants/:id/approve
    → update status: provisioning
    → push tenant ID to provisionQueue (buffered channel)

Goroutine Worker picks job
    → exec provision.sh with tenant credentials
    → stream logs to DB (provision_log)
    → on success: status = active, send ready email
    → on failure: status = failed, log error
```

---

## 2. Go Microservice Project Structure

```
hms-control/
├── main.go
├── go.mod
├── go.sum
├── .env
│
├── cmd/
│   └── migrate/
│       └── main.go               ← run migrations standalone
│
├── internal/
│   ├── config/
│   │   └── config.go             ← env loading with godotenv
│   │
│   ├── db/
│   │   ├── db.go                 ← pgx connection pool
│   │   ├── queries/
│   │   │   ├── tenants.sql       ← raw SQL (used by sqlc)
│   │   │   └── users.sql
│   │   └── generated/            ← sqlc output (do not edit)
│   │       ├── db.go
│   │       ├── models.go
│   │       ├── tenants.sql.go
│   │       └── users.sql.go
│   │
│   ├── handlers/
│   │   ├── auth.go               ← register, login, verify
│   │   ├── admin.go              ← superadmin tenant management
│   │   └── client.go             ← client dashboard
│   │
│   ├── middleware/
│   │   ├── auth.go               ← session/JWT check
│   │   └── role.go               ← superadmin gate
│   │
│   ├── provisioner/
│   │   ├── engine.go             ← goroutine pool + job channel
│   │   ├── runner.go             ← exec provision.sh, stream logs
│   │   └── templates.go          ← generate .env + docker-compose from template
│   │
│   ├── mailer/
│   │   └── mailer.go             ← SMTP via gomail
│   │
│   └── views/                    ← Templ components
│       ├── layout/
│       │   └── base.templ
│       ├── auth/
│       │   ├── register.templ
│       │   └── verify.templ
│       ├── admin/
│       │   ├── tenants_list.templ
│       │   └── tenant_detail.templ
│       └── client/
│           └── dashboard.templ
│
├── docker-templates/
│   ├── docker-compose.template.yml
│   └── .env.template
│
└── scripts/
    ├── provision.sh
    └── update_all.sh
```

---

## 3. Database Design

### Master PostgreSQL Schema

```sql
-- migrations/001_create_users.sql
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    email      VARCHAR(150) NOT NULL UNIQUE,
    company    VARCHAR(200) NOT NULL,
    phone      VARCHAR(25),
    password   VARCHAR(255) NOT NULL,          -- bcrypt
    role       VARCHAR(20)  NOT NULL DEFAULT 'client',  -- client | superadmin
    verified   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- migrations/002_create_tenants.sql
CREATE TYPE tenant_status AS ENUM (
    'pending_verification',
    'pending_approval',
    'provisioning',
    'active',
    'suspended',
    'failed'
);

CREATE TABLE tenants (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name    VARCHAR(200) NOT NULL,
    slug            VARCHAR(50)  NOT NULL UNIQUE,     -- dady2
    domain          VARCHAR(150) NOT NULL UNIQUE,     -- dady2.hms.co.tz
    db_name         VARCHAR(100) NOT NULL UNIQUE,
    db_user         VARCHAR(100) NOT NULL UNIQUE,
    db_password     VARCHAR(100) NOT NULL,
    app_key         VARCHAR(255),
    status          tenant_status NOT NULL DEFAULT 'pending_verification',
    provision_log   TEXT,
    approved_at     TIMESTAMPTZ,
    provisioned_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- migrations/003_create_verify_tokens.sql
CREATE TABLE verify_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN     NOT NULL DEFAULT FALSE
);
```

### Tenant Status Table

| Status | Meaning | Who Triggers |
|---|---|---|
| `pending_verification` | Just registered, email not verified | System on register |
| `pending_approval` | Email verified, waiting on superadmin | Email verify click |
| `provisioning` | Goroutine picked job, Docker running | Superadmin approve |
| `active` | Container live, client can log in | Provisioner on success |
| `suspended` | Manually suspended | Superadmin |
| `failed` | provision.sh exited non-zero | Provisioner on failure |

---

## 4. Go Models & DB Layer (sqlc)

### go.mod

```go
module github.com/dadyutenga/hms-control

go 1.22

require (
    github.com/gofiber/fiber/v2 v2.52.4
    github.com/gofiber/template/html/v2 v2.1.2
    github.com/a-h/templ v0.2.680
    github.com/jackc/pgx/v5 v5.5.5
    github.com/joho/godotenv v1.5.1
    github.com/google/uuid v1.6.0
    golang.org/x/crypto v0.22.0
    gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)
```

### internal/config/config.go

```go
package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    AppURL        string
    BaseDomain    string
    DBUrl         string
    SMTPHost      string
    SMTPPort      string
    SMTPUser      string
    SMTPPass      string
    SMTPFrom      string
    SessionSecret string
    ProvisionScript string
    TenantDir     string
    HMSSource     string
    WorkerCount   int
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file, reading from environment")
    }

    return &Config{
        AppURL:          getEnv("APP_URL",           "https://master.hms.co.tz"),
        BaseDomain:      getEnv("BASE_DOMAIN",       "hms.co.tz"),
        DBUrl:           getEnv("DATABASE_URL",      ""),
        SMTPHost:        getEnv("SMTP_HOST",         ""),
        SMTPPort:        getEnv("SMTP_PORT",         "587"),
        SMTPUser:        getEnv("SMTP_USER",         ""),
        SMTPPass:        getEnv("SMTP_PASS",         ""),
        SMTPFrom:        getEnv("SMTP_FROM",         "noreply@hms.co.tz"),
        SessionSecret:   getEnv("SESSION_SECRET",    "change-me-32-chars-minimum"),
        ProvisionScript: getEnv("PROVISION_SCRIPT",  "/opt/hms-control/scripts/provision.sh"),
        TenantDir:       getEnv("TENANT_DIR",        "/opt/tenants"),
        HMSSource:       getEnv("HMS_SOURCE",        "/opt/hms-source"),
        WorkerCount:     3, // parallel provision goroutines
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

### internal/db/queries/tenants.sql

```sql
-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: ListTenants :many
SELECT t.*, u.name as user_name, u.email as user_email
FROM tenants t
JOIN users u ON t.user_id = u.id
ORDER BY
    CASE t.status
        WHEN 'pending_approval' THEN 1
        WHEN 'provisioning'     THEN 2
        WHEN 'failed'           THEN 3
        WHEN 'active'           THEN 4
        ELSE 5
    END,
    t.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateTenant :one
INSERT INTO tenants (
    user_id, company_name, slug, domain,
    db_name, db_user, db_password
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateTenantStatus :exec
UPDATE tenants
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: ApproveTenant :exec
UPDATE tenants
SET status = 'provisioning', approved_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: SetTenantActive :exec
UPDATE tenants
SET status = 'active',
    app_key = $2,
    provision_log = $3,
    provisioned_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: SetTenantFailed :exec
UPDATE tenants
SET status = 'failed',
    provision_log = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: SetTenantProvisioning :exec
UPDATE tenants
SET status = 'provisioning',
    provision_log = NULL,
    updated_at = NOW()
WHERE id = $1;
```

### internal/db/queries/users.sql

```sql
-- name: CreateUser :one
INSERT INTO users (name, email, company, phone, password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: VerifyUser :exec
UPDATE users SET verified = TRUE, updated_at = NOW() WHERE id = $1;

-- name: CreateVerifyToken :one
INSERT INTO verify_tokens (user_id, token, expires_at)
VALUES ($1, $2, NOW() + INTERVAL '24 hours')
RETURNING *;

-- name: GetVerifyToken :one
SELECT vt.*, u.id as uid FROM verify_tokens vt
JOIN users u ON vt.user_id = u.id
WHERE vt.token = $1 AND vt.used = FALSE AND vt.expires_at > NOW();

-- name: UseVerifyToken :exec
UPDATE verify_tokens SET used = TRUE WHERE token = $1;
```

---

## 5. Fiber App Bootstrap

### main.go

```go
package main

import (
    "log"

    "github.com/dadyutenga/hms-control/internal/config"
    "github.com/dadyutenga/hms-control/internal/db"
    "github.com/dadyutenga/hms-control/internal/handlers"
    "github.com/dadyutenga/hms-control/internal/middleware"
    "github.com/dadyutenga/hms-control/internal/provisioner"
    "github.com/dadyutenga/hms-control/internal/mailer"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/gofiber/fiber/v2/middleware/session"
)

func main() {
    cfg  := config.Load()
    pool := db.Connect(cfg.DBUrl)
    mail := mailer.New(cfg)

    // Start provisioner goroutine pool
    eng := provisioner.NewEngine(cfg, pool, mail)
    eng.Start()

    // Session store
    store := session.New(session.Config{
        KeyLookup: "cookie:hms_session",
        CookieSecure: true,
    })

    app := fiber.New(fiber.Config{
        ErrorHandler: handlers.ErrorHandler,
    })

    app.Use(logger.New())
    app.Use(recover.New())
    app.Static("/static", "./static")

    h := handlers.New(cfg, pool, mail, store, eng)

    // ── Public routes ─────────────────────────────────────────────────────────
    app.Get("/",              h.Home)
    app.Get("/register",      h.ShowRegister)
    app.Post("/register",     h.Register)
    app.Get("/verify/:token", h.VerifyEmail)
    app.Get("/login",         h.ShowLogin)
    app.Post("/login",        h.Login)
    app.Post("/logout",       h.Logout)

    // ── Client routes ─────────────────────────────────────────────────────────
    client := app.Group("/dashboard", middleware.Auth(store, pool))
    client.Get("/", h.ClientDashboard)

    // ── Superadmin routes ─────────────────────────────────────────────────────
    admin := app.Group("/admin",
        middleware.Auth(store, pool),
        middleware.RequireRole("superadmin"),
    )
    admin.Get("/",                         h.AdminDashboard)
    admin.Get("/tenants",                  h.ListTenants)
    admin.Get("/tenants/:id",              h.ShowTenant)
    admin.Post("/tenants/:id/approve",     h.ApproveTenant)
    admin.Post("/tenants/:id/suspend",     h.SuspendTenant)
    admin.Post("/tenants/:id/retry",       h.RetryProvision)

    log.Fatal(app.Listen(":8080"))
}
```

---

## 6. Auth — Registration & Email Verification

### internal/handlers/auth.go

```go
package handlers

import (
    "crypto/rand"
    "encoding/hex"
    "strings"

    "github.com/dadyutenga/hms-control/internal/db/generated"
    "github.com/dadyutenga/hms-control/internal/views/auth"

    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
)

func (h *Handler) ShowRegister(c *fiber.Ctx) error {
    return auth.Register(auth.RegisterProps{}).Render(c.Context(), c.Response().BodyWriter())
}

func (h *Handler) Register(c *fiber.Ctx) error {
    name    := strings.TrimSpace(c.FormValue("name"))
    company := strings.TrimSpace(c.FormValue("company_name"))
    email   := strings.TrimSpace(c.FormValue("email"))
    phone   := strings.TrimSpace(c.FormValue("phone"))
    pass    := c.FormValue("password")
    confirm := c.FormValue("password_confirmation")

    // Basic validation
    if name == "" || company == "" || email == "" || pass == "" {
        return renderRegisterError(c, "All fields are required.")
    }
    if pass != confirm {
        return renderRegisterError(c, "Passwords do not match.")
    }
    if len(pass) < 8 {
        return renderRegisterError(c, "Password must be at least 8 characters.")
    }

    hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

    q := generated.New(h.pool)

    user, err := q.CreateUser(c.Context(), generated.CreateUserParams{
        Name:     name,
        Email:    email,
        Company:  company,
        Phone:    nullString(phone),
        Password: string(hash),
    })
    if err != nil {
        if strings.Contains(err.Error(), "unique") {
            return renderRegisterError(c, "Email already registered.")
        }
        return err
    }

    // Generate tenant slug + DB credentials
    slug    := generateSlug(q, c.Context(), company)
    domain  := slug + "." + h.cfg.BaseDomain
    dbPass  := randomHex(16)

    _, err = q.CreateTenant(c.Context(), generated.CreateTenantParams{
        UserID:      user.ID,
        CompanyName: company,
        Slug:        slug,
        Domain:      domain,
        DbName:      "hms_" + slug + "_db",
        DbUser:      "hms_" + slug + "_user",
        DbPassword:  dbPass,
    })
    if err != nil {
        return err
    }

    // Generate verify token
    tokenBytes := make([]byte, 32)
    rand.Read(tokenBytes)
    token := hex.EncodeToString(tokenBytes)

    q.CreateVerifyToken(c.Context(), generated.CreateVerifyTokenParams{
        UserID: user.ID,
        Token:  token,
    })

    // Send verification email (non-blocking)
    go h.mail.SendVerification(user.Email, user.Name, h.cfg.AppURL+"/verify/"+token)

    return c.Redirect("/verify-notice")
}

func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
    token := c.Params("token")
    q     := generated.New(h.pool)

    row, err := q.GetVerifyToken(c.Context(), token)
    if err != nil {
        return c.Status(400).SendString("Invalid or expired verification link.")
    }

    q.VerifyUser(c.Context(), row.Uid)
    q.UseVerifyToken(c.Context(), token)

    // Update tenant status to pending_approval
    tenant, _ := q.GetTenantByUserID(c.Context(), row.Uid)
    q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
        ID:     tenant.ID,
        Status: "pending_approval",
    })

    return c.Redirect("/login?verified=1")
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func generateSlug(q *generated.Queries, ctx interface{ Value(any) any }, company string) string {
    // strip non-alphanumeric, lowercase, max 12 chars
    re  := strings.NewReplacer(" ", "", "-", "", "_", "")
    base := strings.ToLower(re.Replace(company))
    if len(base) > 12 { base = base[:12] }

    slug := base
    i    := 2
    for {
        _, err := q.GetTenantBySlug(ctx.(fiber.Ctx).Context(), slug)
        if err != nil { break } // slug is free
        slug = base[:10] + string(rune('0'+i))
        i++
    }
    return slug
}

func randomHex(n int) string {
    b := make([]byte, n)
    rand.Read(b)
    return hex.EncodeToString(b)
}

func nullString(s string) interface{} {
    if s == "" { return nil }
    return s
}
```

---

## 7. Provisioner — Core Engine

### internal/provisioner/engine.go

```go
package provisioner

import (
    "context"
    "log"

    "github.com/dadyutenga/hms-control/internal/config"
    "github.com/dadyutenga/hms-control/internal/db/generated"
    "github.com/dadyutenga/hms-control/internal/mailer"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Engine struct {
    cfg   *config.Config
    pool  *pgxpool.Pool
    mail  *mailer.Mailer
    queue chan uuid.UUID  // tenant IDs to provision
}

func NewEngine(cfg *config.Config, pool *pgxpool.Pool, mail *mailer.Mailer) *Engine {
    return &Engine{
        cfg:   cfg,
        pool:  pool,
        mail:  mail,
        queue: make(chan uuid.UUID, 50), // buffer 50 jobs
    }
}

// Start launches N worker goroutines
func (e *Engine) Start() {
    for i := range e.cfg.WorkerCount {
        go e.worker(i)
    }
    log.Printf("[provisioner] started %d workers", e.cfg.WorkerCount)
}

// Enqueue pushes a tenant ID into the provision queue
func (e *Engine) Enqueue(tenantID uuid.UUID) {
    e.queue <- tenantID
}

func (e *Engine) worker(id int) {
    log.Printf("[worker-%d] ready", id)
    for tenantID := range e.queue {
        log.Printf("[worker-%d] provisioning tenant %s", id, tenantID)
        e.provision(tenantID)
    }
}

func (e *Engine) provision(tenantID uuid.UUID) {
    ctx := context.Background()
    q   := generated.New(e.pool)

    tenant, err := q.GetTenantByID(ctx, tenantID)
    if err != nil {
        log.Printf("[provisioner] tenant %s not found: %v", tenantID, err)
        return
    }

    runner := NewRunner(e.cfg)
    logOutput, err := runner.Run(tenant)

    if err != nil {
        log.Printf("[provisioner] FAILED tenant %s: %v", tenant.Slug, err)
        q.SetTenantFailed(ctx, generated.SetTenantFailedParams{
            ID:           tenantID,
            ProvisionLog: &logOutput,
        })
        return
    }

    // Get the generated APP_KEY
    appKey, _ := runner.GetAppKey(tenant.Slug)

    q.SetTenantActive(ctx, generated.SetTenantActiveParams{
        ID:           tenantID,
        AppKey:       &appKey,
        ProvisionLog: &logOutput,
    })

    // Get user email to notify
    user, _ := q.GetUserByID(ctx, tenant.UserID)
    go e.mail.SendTenantReady(user.Email, user.Name, "https://"+tenant.Domain)

    log.Printf("[provisioner] SUCCESS tenant %s live at https://%s", tenant.Slug, tenant.Domain)
}
```

### internal/provisioner/runner.go

```go
package provisioner

import (
    "bytes"
    "fmt"
    "os/exec"
    "strings"

    "github.com/dadyutenga/hms-control/internal/config"
    "github.com/dadyutenga/hms-control/internal/db/generated"
)

type Runner struct {
    cfg *config.Config
}

func NewRunner(cfg *config.Config) *Runner {
    return &Runner{cfg: cfg}
}

// Run executes provision.sh and returns combined stdout+stderr log
func (r *Runner) Run(t generated.Tenant) (string, error) {
    args := []string{
        t.Slug,
        t.Domain,
        t.DbName,
        t.DbUser,
        t.DbPassword,
        r.cfg.TenantDir,
        r.cfg.HMSSource,
    }

    cmd := exec.Command("sudo", append([]string{"bash", r.cfg.ProvisionScript}, args...)...)

    var buf bytes.Buffer
    cmd.Stdout = &buf
    cmd.Stderr = &buf

    err := cmd.Run()
    return buf.String(), err
}

// GetAppKey reads the APP_KEY from a running container
func (r *Runner) GetAppKey(slug string) (string, error) {
    out, err := exec.Command(
        "docker", "exec", "hms_"+slug,
        "php", "artisan", "key:generate", "--show",
    ).Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}
```

### internal/provisioner/templates.go

```go
package provisioner

import (
    "fmt"
    "os"
    "strings"
)

type TenantVars struct {
    Slug    string
    Domain  string
    DBName  string
    DBUser  string
    DBPass  string
}

func GenerateEnv(dir string, v TenantVars) error {
    content := fmt.Sprintf(`APP_NAME="HMS - %s"
APP_ENV=production
APP_KEY=
APP_DEBUG=false
APP_URL=https://%s

DB_CONNECTION=pgsql
DB_HOST=postgres_%s
DB_PORT=5432
DB_DATABASE=%s
DB_USERNAME=%s
DB_PASSWORD=%s

CACHE_DRIVER=file
SESSION_DRIVER=file
QUEUE_CONNECTION=sync
`, v.Slug, v.Domain, v.Slug, v.DBName, v.DBUser, v.DBPass)

    return os.WriteFile(dir+"/.env", []byte(content), 0600)
}

func GenerateCompose(templatePath, outputPath string, v TenantVars) error {
    raw, err := os.ReadFile(templatePath)
    if err != nil {
        return err
    }

    result := strings.NewReplacer(
        "{{SLUG}}",    v.Slug,
        "{{DOMAIN}}",  v.Domain,
        "{{DB_NAME}}", v.DBName,
        "{{DB_USER}}", v.DBUser,
        "{{DB_PASS}}", v.DBPass,
    ).Replace(string(raw))

    return os.WriteFile(outputPath, []byte(result), 0644)
}
```

---

## 8. Provisioner Worker (Goroutine Pool)

The engine uses a **buffered channel** as a job queue and **N goroutines** as workers. This means:

- Multiple tenants can be provisioned in parallel
- No external queue dependency (no Redis, no RabbitMQ)
- Jobs survive handler return — goroutines run independently
- Buffer of 50 means 50 pending provisions before backpressure

```
Approve HTTP handler
    │
    └── eng.Enqueue(tenantID)   ← non-blocking push to channel
            │
            ▼
    chan uuid.UUID (buffer: 50)
            │
    ┌───────┴───────┐
    ▼               ▼
worker-0         worker-1         worker-2
(provision)      (provision)      (idle)
```

To change concurrency — just update `WorkerCount` in config. No other changes needed.

---

## 9. Superadmin Handlers

### internal/handlers/admin.go

```go
package handlers

import (
    "github.com/dadyutenga/hms-control/internal/db/generated"
    "github.com/dadyutenga/hms-control/internal/views/admin"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

func (h *Handler) ListTenants(c *fiber.Ctx) error {
    q       := generated.New(h.pool)
    tenants, err := q.ListTenants(c.Context(), generated.ListTenantsParams{
        Limit:  20,
        Offset: 0,
    })
    if err != nil {
        return err
    }

    // HTMX partial swap vs full page
    if c.Get("HX-Request") == "true" {
        return admin.TenantTable(tenants).Render(c.Context(), c.Response().BodyWriter())
    }
    return admin.TenantList(tenants).Render(c.Context(), c.Response().BodyWriter())
}

func (h *Handler) ApproveTenant(c *fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).SendString("Invalid tenant ID")
    }

    q      := generated.New(h.pool)
    tenant, err := q.GetTenantByID(c.Context(), id)
    if err != nil {
        return fiber.ErrNotFound
    }
    if tenant.Status != "pending_approval" {
        return c.Status(409).SendString("Tenant not in pending_approval state")
    }

    q.ApproveTenant(c.Context(), id)

    // Push to provisioner engine — non-blocking
    h.eng.Enqueue(id)

    // HTMX swap the status badge inline
    if c.Get("HX-Request") == "true" {
        return admin.StatusBadge("provisioning").Render(c.Context(), c.Response().BodyWriter())
    }
    return c.Redirect("/admin/tenants")
}

func (h *Handler) SuspendTenant(c *fiber.Ctx) error {
    id, _ := uuid.Parse(c.Params("id"))
    q     := generated.New(h.pool)
    q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
        ID: id, Status: "suspended",
    })

    if c.Get("HX-Request") == "true" {
        return admin.StatusBadge("suspended").Render(c.Context(), c.Response().BodyWriter())
    }
    return c.Redirect("/admin/tenants")
}

func (h *Handler) RetryProvision(c *fiber.Ctx) error {
    id, _ := uuid.Parse(c.Params("id"))
    q     := generated.New(h.pool)

    tenant, _ := q.GetTenantByID(c.Context(), id)
    if tenant.Status != "failed" {
        return c.Status(409).SendString("Only failed tenants can be retried")
    }

    q.SetTenantProvisioning(c.Context(), id)
    h.eng.Enqueue(id)

    if c.Get("HX-Request") == "true" {
        return admin.StatusBadge("provisioning").Render(c.Context(), c.Response().BodyWriter())
    }
    return c.Redirect("/admin/tenants/"+id.String())
}
```

---

## 10. Templ + HTMX Views

### internal/views/layout/base.templ

```go
package layout

templ Base(title string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
        <title>{ title } — HMS Control</title>
        <script src="https://unpkg.com/htmx.org@1.9.12"></script>
        <link rel="stylesheet" href="/static/css/app.css"/>
    </head>
    <body class="bg-slate-950 text-slate-100 min-h-screen">
        { children... }
    </body>
    </html>
}
```

### internal/views/admin/tenants_list.templ

```go
package admin

import (
    "github.com/dadyutenga/hms-control/internal/db/generated"
)

templ TenantList(tenants []generated.ListTenantsRow) {
    @layout.Base("Tenants") {
        <div class="max-w-7xl mx-auto p-6">
            <div class="flex justify-between items-center mb-6">
                <h1 class="text-2xl font-bold text-white">Tenants</h1>
                <span class="text-slate-400">{ len(tenants) } total</span>
            </div>
            @TenantTable(tenants)
        </div>
    }
}

templ TenantTable(tenants []generated.ListTenantsRow) {
    <div id="tenant-table">
        <table class="w-full text-sm">
            <thead>
                <tr class="border-b border-slate-800 text-slate-400">
                    <th class="text-left py-3">Company</th>
                    <th class="text-left py-3">Domain</th>
                    <th class="text-left py-3">Status</th>
                    <th class="text-left py-3">Registered</th>
                    <th class="text-left py-3">Actions</th>
                </tr>
            </thead>
            <tbody>
                for _, t := range tenants {
                    <tr class="border-b border-slate-800/50 hover:bg-slate-900/50">
                        <td class="py-3">
                            <div class="font-medium">{ t.CompanyName }</div>
                            <div class="text-slate-500 text-xs">{ t.UserEmail }</div>
                        </td>
                        <td class="py-3">
                            <a href={ templ.URL("https://" + t.Domain) }
                               class="text-indigo-400 hover:underline" target="_blank">
                                { t.Domain }
                            </a>
                        </td>
                        <td class="py-3" id={ "status-" + t.ID.String() }>
                            @StatusBadge(string(t.Status))
                        </td>
                        <td class="py-3 text-slate-400">
                            { t.CreatedAt.Time.Format("02 Jan 2006") }
                        </td>
                        <td class="py-3">
                            @TenantActions(t)
                        </td>
                    </tr>
                }
            </tbody>
        </table>
    </div>
}

templ TenantActions(t generated.ListTenantsRow) {
    <div class="flex gap-2">
        if t.Status == "pending_approval" {
            <button
                hx-post={ "/admin/tenants/" + t.ID.String() + "/approve" }
                hx-target={ "#status-" + t.ID.String() }
                hx-swap="innerHTML"
                hx-confirm="Approve and provision this tenant?"
                class="px-3 py-1 bg-indigo-600 hover:bg-indigo-500 rounded text-xs font-medium">
                Approve
            </button>
        }
        if t.Status == "active" {
            <button
                hx-post={ "/admin/tenants/" + t.ID.String() + "/suspend" }
                hx-target={ "#status-" + t.ID.String() }
                hx-swap="innerHTML"
                class="px-3 py-1 bg-red-700 hover:bg-red-600 rounded text-xs font-medium">
                Suspend
            </button>
        }
        if t.Status == "failed" {
            <button
                hx-post={ "/admin/tenants/" + t.ID.String() + "/retry" }
                hx-target={ "#status-" + t.ID.String() }
                hx-swap="innerHTML"
                class="px-3 py-1 bg-orange-600 hover:bg-orange-500 rounded text-xs font-medium">
                Retry
            </button>
        }
        <a href={ templ.URL("/admin/tenants/" + t.ID.String()) }
           class="px-3 py-1 bg-slate-700 hover:bg-slate-600 rounded text-xs font-medium">
            Logs
        </a>
    </div>
}

templ StatusBadge(status string) {
    switch status {
        case "pending_verification":
            <span class="px-2 py-1 bg-slate-700 text-slate-300 rounded text-xs">unverified</span>
        case "pending_approval":
            <span class="px-2 py-1 bg-purple-900 text-purple-300 rounded text-xs">pending</span>
        case "provisioning":
            <span class="px-2 py-1 bg-cyan-900 text-cyan-300 rounded text-xs animate-pulse">provisioning</span>
        case "active":
            <span class="px-2 py-1 bg-green-900 text-green-300 rounded text-xs">active</span>
        case "suspended":
            <span class="px-2 py-1 bg-red-900 text-red-300 rounded text-xs">suspended</span>
        case "failed":
            <span class="px-2 py-1 bg-orange-900 text-orange-300 rounded text-xs">failed</span>
        default:
            <span class="px-2 py-1 bg-slate-700 text-slate-400 rounded text-xs">{ status }</span>
    }
}
```

---

## 11. Email Service

### internal/mailer/mailer.go

```go
package mailer

import (
    "fmt"

    "github.com/dadyutenga/hms-control/internal/config"
    "gopkg.in/gomail.v2"
    "strconv"
)

type Mailer struct {
    cfg *config.Config
}

func New(cfg *config.Config) *Mailer {
    return &Mailer{cfg: cfg}
}

func (m *Mailer) send(to, subject, body string) error {
    port, _ := strconv.Atoi(m.cfg.SMTPPort)
    d := gomail.NewDialer(m.cfg.SMTPHost, port, m.cfg.SMTPUser, m.cfg.SMTPPass)

    msg := gomail.NewMessage()
    msg.SetHeader("From", m.cfg.SMTPFrom)
    msg.SetHeader("To", to)
    msg.SetHeader("Subject", subject)
    msg.SetBody("text/html", body)

    return d.DialAndSend(msg)
}

func (m *Mailer) SendVerification(to, name, verifyURL string) error {
    body := fmt.Sprintf(`
        <h2>Hello %s,</h2>
        <p>Thanks for registering on HMS Platform.</p>
        <p>Please verify your email address:</p>
        <a href="%s" style="background:#6366f1;color:white;padding:12px 24px;
           border-radius:6px;text-decoration:none;display:inline-block;">
            Verify Email
        </a>
        <p>This link expires in 24 hours.</p>
    `, name, verifyURL)

    return m.send(to, "Verify Your Email — HMS Platform", body)
}

func (m *Mailer) SendTenantReady(to, name, hmsURL string) error {
    body := fmt.Sprintf(`
        <h2>Hello %s,</h2>
        <p>Your Hotel Management System is live!</p>
        <a href="%s" style="background:#22c55e;color:white;padding:12px 24px;
           border-radius:6px;text-decoration:none;display:inline-block;">
            Access Your HMS
        </a>
        <p>Login with the credentials you registered with.</p>
        <p>Need help? Email support@hms.co.tz</p>
    `, name, hmsURL)

    return m.send(to, "Your HMS is Ready!", body)
}
```

---

## 12. Docker Templates

### docker-templates/docker-compose.template.yml

```yaml
version: '3.8'

services:
  app_{{SLUG}}:
    build:
      context: /opt/hms-source
      dockerfile: Dockerfile
    image: hms-app:latest
    container_name: hms_{{SLUG}}
    restart: unless-stopped
    env_file: /opt/tenants/{{SLUG}}/.env
    volumes:
      - /opt/tenants/{{SLUG}}/storage:/var/www/html/storage
    networks:
      - traefik_net
    depends_on:
      - postgres_{{SLUG}}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.{{SLUG}}.rule=Host(`{{DOMAIN}}`)"
      - "traefik.http.routers.{{SLUG}}.entrypoints=websecure"
      - "traefik.http.routers.{{SLUG}}.tls.certresolver=letsencrypt"
      - "traefik.http.services.{{SLUG}}.loadbalancer.server.port=9000"

  postgres_{{SLUG}}:
    image: postgres:15-alpine
    container_name: postgres_{{SLUG}}
    restart: unless-stopped
    environment:
      POSTGRES_DB: {{DB_NAME}}
      POSTGRES_USER: {{DB_USER}}
      POSTGRES_PASSWORD: {{DB_PASS}}
    volumes:
      - pgdata_{{SLUG}}:/var/lib/postgresql/data
    networks:
      - traefik_net

volumes:
  pgdata_{{SLUG}}:

networks:
  traefik_net:
    external: true
```

---

## 13. Provision Shell Script

### scripts/provision.sh

```bash
#!/bin/bash
set -euo pipefail

SLUG=$1
DOMAIN=$2
DB_NAME=$3
DB_USER=$4
DB_PASS=$5
TENANT_DIR=$6   # /opt/tenants
HMS_SOURCE=$7   # /opt/hms-source

DIR="${TENANT_DIR}/${SLUG}"

echo "[provision] Starting tenant: ${SLUG}"
echo "[provision] Domain:          ${DOMAIN}"

# 1. Create directories
mkdir -p "${DIR}/storage/app/public"
mkdir -p "${DIR}/storage/framework/cache"
mkdir -p "${DIR}/storage/framework/sessions"
mkdir -p "${DIR}/storage/framework/views"
mkdir -p "${DIR}/storage/logs"

# 2. Generate .env
cat > "${DIR}/.env" <<EOF
APP_NAME="HMS - ${SLUG}"
APP_ENV=production
APP_KEY=
APP_DEBUG=false
APP_URL=https://${DOMAIN}

DB_CONNECTION=pgsql
DB_HOST=postgres_${SLUG}
DB_PORT=5432
DB_DATABASE=${DB_NAME}
DB_USERNAME=${DB_USER}
DB_PASSWORD=${DB_PASS}

CACHE_DRIVER=file
SESSION_DRIVER=file
QUEUE_CONNECTION=sync
EOF

# 3. Generate docker-compose from template
TEMPLATE="/opt/hms-control/docker-templates/docker-compose.template.yml"
sed \
  -e "s|{{SLUG}}|${SLUG}|g" \
  -e "s|{{DOMAIN}}|${DOMAIN}|g" \
  -e "s|{{DB_NAME}}|${DB_NAME}|g" \
  -e "s|{{DB_USER}}|${DB_USER}|g" \
  -e "s|{{DB_PASS}}|${DB_PASS}|g" \
  "${TEMPLATE}" > "${DIR}/docker-compose.yml"

# 4. Spin up containers
cd "${DIR}"
docker compose up -d --build

# 5. Wait for PostgreSQL
echo "[provision] Waiting for PostgreSQL..."
for i in $(seq 1 30); do
    docker exec "postgres_${SLUG}" pg_isready -U "${DB_USER}" && break
    sleep 2
done

# 6. Laravel setup
docker exec "hms_${SLUG}" php artisan key:generate --force
docker exec "hms_${SLUG}" php artisan migrate --force
docker exec "hms_${SLUG}" php artisan db:seed --class=DefaultDataSeeder --force
docker exec "hms_${SLUG}" php artisan config:cache
docker exec "hms_${SLUG}" php artisan route:cache
docker exec "hms_${SLUG}" php artisan storage:link

echo "[provision] DONE — https://${DOMAIN} is live"
```

---

## 14. Traefik & DNS

### /opt/traefik/docker-compose.yml

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v3.0
    container_name: traefik
    restart: unless-stopped
    command:
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--entrypoints.web.http.redirections.entrypoint.to=websecure"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@hms.co.tz"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - letsencrypt:/letsencrypt
    networks:
      - traefik_net

volumes:
  letsencrypt:

networks:
  traefik_net:
    name: traefik_net
    driver: bridge
```

### DNS Records

| Type | Name | Value |
|---|---|---|
| A | `*.hms.co.tz` | `YOUR_SERVER_IP` |
| A | `hms.co.tz` | `YOUR_SERVER_IP` |
| A | `master.hms.co.tz` | `YOUR_SERVER_IP` |

---

## 15. Deployment & Server Setup

### .env for Go control plane

```env
APP_URL=https://master.hms.co.tz
BASE_DOMAIN=hms.co.tz
DATABASE_URL=postgres://hmsadmin:secret@localhost:5432/hms_master?sslmode=disable

SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USER=postmaster@hms.co.tz
SMTP_PASS=your-smtp-password
SMTP_FROM=noreply@hms.co.tz

SESSION_SECRET=your-32-char-secret-here

PROVISION_SCRIPT=/opt/hms-control/scripts/provision.sh
TENANT_DIR=/opt/tenants
HMS_SOURCE=/opt/hms-source
```

### Step-by-Step Bootstrap

```bash
# 1. Install Go 1.22+
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 2. Install Docker
curl -fsSL https://get.docker.com | sh

# 3. Create directories
sudo mkdir -p /opt/hms-control /opt/hms-source /opt/tenants /opt/traefik

# 4. Clone and build the Go control plane
cd /opt/hms-control
git clone <control-repo> .
go build -o hms-control ./cmd/...

# 5. Clone HMS source (Laravel)
cd /opt/hms-source
git clone <hms-repo> .

# 6. Run DB migrations
./hms-control migrate

# 7. Seed superadmin
./hms-control seed:admin

# 8. Make scripts executable
chmod +x /opt/hms-control/scripts/provision.sh

# 9. Sudoers for provision script
echo "hmsctl ALL=(ALL) NOPASSWD: /bin/bash /opt/hms-control/scripts/provision.sh" \
  | sudo tee /etc/sudoers.d/hms-provision

# 10. Start Traefik
cd /opt/traefik
docker network create traefik_net
docker compose up -d

# 11. Run HMS control plane as systemd service (see below)
```

### systemd — /etc/systemd/system/hms-control.service

```ini
[Unit]
Description=HMS Control Plane (Go/Fiber)
After=network.target postgresql.service

[Service]
User=hmsctl
WorkingDirectory=/opt/hms-control
ExecStart=/opt/hms-control/hms-control
Restart=always
RestartSec=3
EnvironmentFile=/opt/hms-control/.env
StandardOutput=journal
StandardError=journal
SyslogIdentifier=hms-control

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable hms-control
sudo systemctl start hms-control
sudo journalctl -u hms-control -f
```

### Final Deployment Checklist

| Item | Command to verify |
|---|---|
| `[ ]` Go binary built | `./hms-control --version` |
| `[ ]` PostgreSQL master DB ready | `psql $DATABASE_URL -c '\dt'` |
| `[ ]` Migrations ran | `./hms-control migrate:status` |
| `[ ]` Traefik running | `docker ps \| grep traefik` |
| `[ ]` traefik_net created | `docker network ls \| grep traefik` |
| `[ ]` Wildcard DNS live | `dig dady2.hms.co.tz +short` |
| `[ ]` hms-control service running | `systemctl status hms-control` |
| `[ ]` Superadmin seeded | Login at `master.hms.co.tz/login` |
| `[ ]` provision.sh executable | `ls -la /opt/hms-control/scripts/` |
| `[ ]` sudoers in place | `sudo cat /etc/sudoers.d/hms-provision` |
| `[ ]` SMTP working | Test via register flow |

---

## 16. Updating All Tenants

When a new version of the Laravel HMS ships:

```bash
#!/bin/bash
# scripts/update_all.sh

echo "=== Rebuilding HMS image ==="
docker build -t hms-app:latest /opt/hms-source

echo "=== Rolling update across all tenants ==="
for dir in /opt/tenants/*/; do
    slug=$(basename "$dir")
    echo "  → ${slug}"
    cd "$dir"
    docker compose up -d --build
    docker exec "hms_${slug}" php artisan migrate --force
    docker exec "hms_${slug}" php artisan config:cache
done

echo "=== All tenants updated ==="
```

Add a handler in Go to trigger this from the superadmin dashboard:

```go
// POST /admin/update-all
func (h *Handler) UpdateAllTenants(c *fiber.Ctx) error {
    go func() {
        out, err := exec.Command("sudo", "bash",
            "/opt/hms-control/scripts/update_all.sh").CombinedOutput()
        if err != nil {
            log.Printf("[update-all] FAILED: %v\n%s", err, out)
            return
        }
        log.Printf("[update-all] done\n%s", out)
    }()

    if c.Get("HX-Request") == "true" {
        return c.SendString(`<span class="text-green-400">Update triggered in background</span>`)
    }
    return c.Redirect("/admin")
}
```

---

## Summary: Why This Stack Wins

| Layer | Tech | Why |
|---|---|---|
| Control plane | Go + Fiber | Single binary, 5ms startup, goroutine-native concurrency |
| Views | Templ + HTMX | Type-safe HTML, no JS build step, partial swaps free |
| Provisioning | Goroutine pool + buffered channel | No Redis needed, N parallel provisions out of the box |
| Tenant app | Laravel (unchanged) | Your existing work stays, zero migration cost |
| Routing | Traefik | Auto SSL per subdomain, Docker-label driven |
| DB | PostgreSQL | One master DB for control plane, isolated DB per tenant |

---

*HMS Master SaaS Platform — Built by Dady Utenga · BIG LITE CODE · Iringa, Tanzania*
*github.com/dadyutenga · dadyprojects.tech · @DadyUtenga*
