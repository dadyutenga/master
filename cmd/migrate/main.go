package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/dadyutenga/hms-control/internal/config"
	"github.com/dadyutenga/hms-control/internal/db"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate <up|down|seed>")
	}

	cfg := config.Load()
	action := os.Args[1]

	database := db.Connect(cfg.DBPath)
	defer database.Close()

	switch action {
	case "up":
		runMigrationsUp(database)
	case "down":
		runMigrationsDown(database)
	case "seed":
		runMigrationsUp(database)
		seedSuperadmin(database, cfg)
	default:
		log.Fatalf("Unknown action: %s (use 'up', 'down', or 'seed')", action)
	}
}

func runMigrationsUp(db *sql.DB) {
	migrations := []string{
		migrationUsers,
		migrationTenants,
	}
	for i, m := range migrations {
		fmt.Printf("Running migration %d...\n", i+1)
		_, err := db.Exec(m)
		if err != nil {
			log.Fatalf("Migration %d failed: %v", i+1, err)
		}
	}

	// Add columns that may be missing from existing tables (ignore errors)
	alterColumns := []string{
		"ALTER TABLE users ADD COLUMN tin TEXT",
		"ALTER TABLE users ADD COLUMN brela_number TEXT",
		"ALTER TABLE tenants ADD COLUMN hotel_name TEXT",
		"ALTER TABLE tenants ADD COLUMN category TEXT",
		"ALTER TABLE tenants ADD COLUMN room_count INTEGER",
		"ALTER TABLE tenants ADD COLUMN address TEXT",
		"ALTER TABLE tenants ADD COLUMN city TEXT",
		"ALTER TABLE tenants ADD COLUMN country TEXT",
		"ALTER TABLE tenants ADD COLUMN requested_subdomain TEXT",
		"ALTER TABLE tenants ADD COLUMN admin_name TEXT",
		"ALTER TABLE tenants ADD COLUMN admin_email TEXT",
		"ALTER TABLE tenants ADD COLUMN admin_phone TEXT",
		"ALTER TABLE tenants ADD COLUMN billing_status TEXT NOT NULL DEFAULT 'unpaid'",
		"ALTER TABLE tenants ADD COLUMN last_payment_at DATETIME",
		"ALTER TABLE tenants ADD COLUMN next_due_at DATETIME",
	}
	for _, a := range alterColumns {
		db.Exec(a)
	}

	// Create index on requested_subdomain (now that column exists)
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_requested_subdomain ON tenants(requested_subdomain)")

	// Rename superadmin -> admin (idempotent)
	db.Exec("UPDATE users SET role = 'admin' WHERE role = 'superadmin'")

	remaining := []string{
		migrationVerifyTokens,
		migrationDocuments,
		migrationContactDetails,
		migrationDeployments,
		migrationAuditLog,
		migrationSettings,
		migrationDockerTemplates,
		migrationBillingTransactions,
		migrationInstances,
		migrationInstanceDeployments,
		migrationBillingPackages,
	}
	for i, m := range remaining {
		fmt.Printf("Running migration %d...\n", i+len(migrations)+1)
		_, err := db.Exec(m)
		if err != nil {
			log.Fatalf("Migration %d failed: %v", i+len(migrations)+1, err)
		}
	}
	fmt.Println("All migrations applied successfully.")

	// Add columns that may be missing from instances table
	db.Exec("ALTER TABLE instances ADD COLUMN slug TEXT NOT NULL DEFAULT ''")

	// Reset billing status: tenants with no actual payments should be 'unpaid'
	db.Exec(`UPDATE tenants SET billing_status = 'unpaid' WHERE billing_status = 'paid' AND id NOT IN (SELECT DISTINCT tenant_id FROM billing_transactions WHERE transaction_type = 'payment' AND status = 'completed')`)
	fmt.Println("Billing status reset for tenants with no payments.")

	// Reset auto-generated domains to pending format
	db.Exec(`UPDATE tenants SET domain = 'pending-' || substr(id, 1, 8) WHERE domain LIKE '%.hms.co.tz'`)
	fmt.Println("Auto-generated domains reset to pending.")
}

func runMigrationsDown(db *sql.DB) {
	drops := []string{
		"DROP TABLE IF EXISTS billing_packages",
		"DROP TABLE IF EXISTS instance_deployments",
		"DROP TABLE IF EXISTS instances",
		"DROP TABLE IF EXISTS docker_templates",
		"DROP TABLE IF EXISTS audit_logs",
		"DROP TABLE IF EXISTS settings",
		"DROP TABLE IF EXISTS deployments",
		"DROP TABLE IF EXISTS contact_details",
		"DROP TABLE IF EXISTS documents",
		"DROP TABLE IF EXISTS verify_tokens",
		"DROP TABLE IF EXISTS tenants",
		"DROP TABLE IF EXISTS users",
	}

	for _, d := range drops {
		_, err := db.Exec(d)
		if err != nil {
			log.Printf("Warning: %v", err)
		}
	}
	fmt.Println("All tables dropped.")
}

func seedSuperadmin(db *sql.DB, cfg *config.Config) {
	if cfg.SuperAdminEmail == "" || cfg.SuperAdminPass == "" || cfg.SuperAdminName == "" {
		fmt.Println("Admin seed skipped: set SUPERADMIN_NAME, SUPERADMIN_EMAIL, SUPERADMIN_PASSWORD")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.SuperAdminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Warning seeding admin: %v", err)
		return
	}
	_, err = db.Exec(
		`INSERT INTO users (name, email, company, password, role, verified)
		 VALUES (?, ?, ?, ?, 'admin', 1)
		 ON CONFLICT(email) DO NOTHING`,
		cfg.SuperAdminName, cfg.SuperAdminEmail, "HMS Platform",
		string(hash),
	)
	if err != nil {
		log.Printf("Warning seeding admin: %v", err)
	} else {
		fmt.Println("Admin seeded. Change the password after first login!")
	}
}

const migrationUsers = `
CREATE TABLE IF NOT EXISTS users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL UNIQUE,
    company    TEXT NOT NULL,
    phone      TEXT,
    password   TEXT NOT NULL,
    role       TEXT NOT NULL DEFAULT 'client',
    verified   INTEGER NOT NULL DEFAULT 0,
    tin        TEXT,
    brela_number TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const migrationTenants = `
CREATE TABLE IF NOT EXISTS tenants (
    id              TEXT PRIMARY KEY,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name    TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    domain          TEXT NOT NULL UNIQUE,
    db_name         TEXT NOT NULL UNIQUE,
    db_user         TEXT NOT NULL UNIQUE,
    db_password     TEXT NOT NULL,
    app_key         TEXT,
    status          TEXT NOT NULL DEFAULT 'pending_verification' CHECK(status IN ('pending_verification','pending_approval','provisioning','active','suspended','failed')),
    provision_log   TEXT,
    approved_at     DATETIME,
    provisioned_at  DATETIME,
    hotel_name      TEXT,
    category        TEXT,
    room_count      INTEGER,
    address         TEXT,
    city            TEXT,
    country         TEXT,
    requested_subdomain TEXT,
    admin_name      TEXT,
    admin_email     TEXT,
    admin_phone     TEXT,
    billing_status  TEXT NOT NULL DEFAULT 'unpaid' CHECK(billing_status IN ('unpaid','paid','overdue','suspended')),
    last_payment_at DATETIME,
    next_due_at     DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const migrationDocuments = `
CREATE TABLE IF NOT EXISTS documents (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id     TEXT REFERENCES tenants(id) ON DELETE CASCADE,
    doc_type      TEXT NOT NULL CHECK(doc_type IN ('brela_certificate','tra_certificate','other')),
    filename      TEXT NOT NULL,
    original_name TEXT NOT NULL,
    mime_type     TEXT NOT NULL,
    size_bytes    INTEGER NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const migrationContactDetails = `
CREATE TABLE IF NOT EXISTS contact_details (
    id           INTEGER PRIMARY KEY,
    location     TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO contact_details (id, location, phone_number, created_at, updated_at)
SELECT 1, 'Samora Avenue, Dar es Salaam', '+255 123 456 789', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM contact_details);
`

const migrationDeployments = `
CREATE TABLE IF NOT EXISTS deployments (
    id            TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    action        TEXT NOT NULL,
    status        TEXT NOT NULL,
    log           TEXT,
    error_message TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at  DATETIME
);
CREATE INDEX IF NOT EXISTS idx_deployments_tenant_id ON deployments(tenant_id);
`

const migrationVerifyTokens = `
CREATE TABLE IF NOT EXISTS verify_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0
);
`

const migrationAuditLog = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id   INTEGER NOT NULL,
    action     TEXT    NOT NULL,
    tenant_id  TEXT,
    detail     TEXT    DEFAULT '',
    ip_address TEXT    DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id  ON audit_logs(tenant_id);
`

const migrationSettings = `
CREATE TABLE IF NOT EXISTS settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT OR IGNORE INTO settings (key, value) VALUES
    ('smtp_host',        ''),
    ('smtp_port',        '587'),
    ('smtp_user',        ''),
    ('smtp_pass',        ''),
    ('smtp_from',        'noreply@localhost'),
    ('provision_script', './scripts/provision.sh'),
    ('docker_template',  'default');
`

const migrationDockerTemplates = `
CREATE TABLE IF NOT EXISTS docker_templates (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL UNIQUE,
    description   TEXT NOT NULL DEFAULT '',
    template_body TEXT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_docker_templates_name ON docker_templates(name);
`

const migrationBillingTransactions = `
CREATE TABLE IF NOT EXISTS billing_transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id       TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    amount          REAL NOT NULL,
    currency        TEXT NOT NULL DEFAULT 'TZS',
    description     TEXT NOT NULL DEFAULT '',
    transaction_type TEXT NOT NULL CHECK(transaction_type IN ('charge','payment','refund','adjustment')),
    status          TEXT NOT NULL DEFAULT 'completed' CHECK(status IN ('pending','completed','failed','refunded')),
    admin_id        INTEGER REFERENCES users(id),
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_billing_txn_tenant_id ON billing_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_billing_txn_created_at ON billing_transactions(created_at DESC);
`

const migrationInstances = `
CREATE TABLE IF NOT EXISTS instances (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    hotel_name      TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    domain          TEXT NOT NULL DEFAULT '',
    db_name         TEXT NOT NULL,
    db_user         TEXT NOT NULL,
    db_password     TEXT NOT NULL,
    app_key         TEXT,
    status          TEXT NOT NULL DEFAULT 'pending_payment'
                        CHECK(status IN ('pending_payment','provisioning','active','paused','disabled','archived','deleted','failed')),
    admin_disabled  INTEGER NOT NULL DEFAULT 0,
    billing_status  TEXT NOT NULL DEFAULT 'unpaid'
                        CHECK(billing_status IN ('unpaid','paid','overdue','suspended')),
    price           REAL NOT NULL DEFAULT 0,
    package_name    TEXT NOT NULL DEFAULT 'basic',
    last_payment_at DATETIME,
    next_due_at     DATETIME,
    provision_log   TEXT,
    approved_at     DATETIME,
    provisioned_at  DATETIME,
    archived_at     DATETIME,
    deleted_at      DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_instances_tenant_id ON instances(tenant_id);
CREATE INDEX IF NOT EXISTS idx_instances_status ON instances(status);
`

const migrationInstanceDeployments = `
CREATE TABLE IF NOT EXISTS instance_deployments (
    id              TEXT PRIMARY KEY,
    instance_id     TEXT NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    action          TEXT NOT NULL,
    status          TEXT NOT NULL,
    log             TEXT,
    error_message   TEXT,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at    DATETIME
);
CREATE INDEX IF NOT EXISTS idx_instance_deployments_instance_id ON instance_deployments(instance_id);
`

const migrationBillingPackages = `
CREATE TABLE IF NOT EXISTS billing_packages (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL UNIQUE,
    description     TEXT NOT NULL DEFAULT '',
    price           REAL NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'TZS',
    billing_cycle   TEXT NOT NULL DEFAULT 'monthly'
                        CHECK(billing_cycle IN ('monthly','quarterly','yearly','one_time')),
    is_active       INTEGER NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT OR IGNORE INTO billing_packages (name, description, price, billing_cycle) VALUES
    ('basic', 'Basic hotel management package', 50000, 'monthly'),
    ('standard', 'Standard hotel management package', 100000, 'monthly'),
    ('premium', 'Premium hotel management package', 200000, 'monthly');
`
