package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/dadyutenga/hms-control/internal/config"
	"github.com/dadyutenga/hms-control/internal/db"

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
		migrationVerifyTokens,
		migrationDocuments,
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
	}
	for _, a := range alterColumns {
		db.Exec(a) // ignore error if column already exists
	}
	fmt.Println("All migrations applied successfully.")
}

func runMigrationsDown(db *sql.DB) {
	drops := []string{
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
	_, err := db.Exec(
		`INSERT INTO users (name, email, company, password, role, verified)
		 VALUES (?, ?, ?, ?, 'superadmin', 1)
		 ON CONFLICT(email) DO NOTHING`,
		"Super Admin", "admin@hms.co.tz", "HMS Platform",
		"$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
	)
	if err != nil {
		log.Printf("Warning seeding superadmin: %v", err)
	} else {
		fmt.Println("Superadmin seeded. Change the password after first login!")
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

const migrationVerifyTokens = `
CREATE TABLE IF NOT EXISTS verify_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0
);
`