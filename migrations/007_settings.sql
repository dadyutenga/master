-- 007_settings.sql
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
    ('smtp_tls',         'true'),
    ('provision_script', './scripts/provision.sh'),
    ('docker_template',  'default'),
    ('provision_timeout','300');
