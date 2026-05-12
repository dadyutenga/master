-- Migration 009: Instance Management
-- Adds instances table for multi-hotel support per tenant

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

-- Instance deployments (per instance, not per tenant)
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

-- Billing packages (superadmin sets pricing per package)
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

-- Insert default packages
INSERT OR IGNORE INTO billing_packages (name, description, price, billing_cycle) VALUES
    ('basic', 'Basic hotel management package', 50000, 'monthly'),
    ('standard', 'Standard hotel management package', 100000, 'monthly'),
    ('premium', 'Premium hotel management package', 200000, 'monthly');
