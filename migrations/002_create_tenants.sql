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
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);