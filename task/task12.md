Task 12 - Settings Expansion (SMTP + Provisioner)

Objective:
Add settings storage and admin pages for SMTP and provisioner configuration.

---

### 1) Migration

Create migrations/007_settings.sql:
```sql
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
```

---

### 2) Routes

Add to main.go:
```go
admin.Get("/settings/smtp",         h.AdminSMTPSettings)
admin.Post("/settings/smtp",        h.UpdateSMTPSettings)
admin.Post("/settings/smtp/test",   h.TestSMTP)
admin.Get("/settings/provisioner",  h.AdminProvisionerSettings)
admin.Post("/settings/provisioner", h.UpdateProvisionerSettings)
```

---

### 3) Helpers

Add to internal/handlers/admin.go:
```go
// getSetting reads one key. Returns "" on miss.
func (h *Handler) getSetting(key string) string {
    var value string
    h.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
    return value
}

// setSetting upserts one key.
func (h *Handler) setSetting(key, value string) error {
    _, err := h.db.Exec(`
        INSERT INTO settings (key, value, updated_at)
        VALUES (?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
    `, key, value)
    return err
}
```

---

### 4) SMTP handlers

AdminSMTPSettings:
```go
func (h *Handler) AdminSMTPSettings(c *fiber.Ctx) error {
    settings := map[string]string{
        "smtp_host": h.getSetting("smtp_host"),
        "smtp_port": h.getSetting("smtp_port"),
        "smtp_user": h.getSetting("smtp_user"),
        "smtp_from": h.getSetting("smtp_from"),
    }
    return render(c, views.SMTPSettingsPage(settings))
}
```

UpdateSMTPSettings:
```go
func (h *Handler) UpdateSMTPSettings(c *fiber.Ctx) error {
    fields := map[string]string{
        "smtp_host": c.FormValue("smtp_host"),
        "smtp_port": c.FormValue("smtp_port"),
        "smtp_user": c.FormValue("smtp_user"),
        "smtp_from": c.FormValue("smtp_from"),
    }
    if pass := c.FormValue("smtp_pass"); pass != "" {
        fields["smtp_pass"] = pass
    }
    for key, value := range fields {
        if err := h.setSetting(key, value); err != nil {
            return err
        }
    }
    sess, _ := h.store.Get(c)
    adminID := sess.Get("user_id").(int64)
    LogAction(h.db, adminID, "settings.smtp_updated", nil, "", c.IP())
    return c.Redirect("/admin/settings/smtp")
}
```

TestSMTP:
```go
func (h *Handler) TestSMTP(c *fiber.Ctx) error {
    to := c.FormValue("test_email")
    if to == "" {
        return c.JSON(fiber.Map{"ok": false, "error": "test_email is required"})
    }
    err := h.mail.Send(to, "HMS SMTP Test", "If you see this, SMTP is configured correctly.")
    if err != nil {
        return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
    }
    return c.JSON(fiber.Map{"ok": true})
}
```

---

### 5) Provisioner handlers

AdminProvisionerSettings:
```go
func (h *Handler) AdminProvisionerSettings(c *fiber.Ctx) error {
    settings := map[string]string{
        "provision_script": h.getSetting("provision_script"),
        "docker_template":  h.getSetting("docker_template"),
    }
    return render(c, views.ProvisionerSettingsPage(settings))
}
```

UpdateProvisionerSettings:
```go
func (h *Handler) UpdateProvisionerSettings(c *fiber.Ctx) error {
    fields := map[string]string{
        "provision_script": c.FormValue("provision_script"),
        "docker_template":  c.FormValue("docker_template"),
    }
    for key, value := range fields {
        if err := h.setSetting(key, value); err != nil {
            return err
        }
    }
    sess, _ := h.store.Get(c)
    adminID := sess.Get("user_id").(int64)
    LogAction(h.db, adminID, "settings.provisioner_updated", nil, "", c.IP())
    return c.Redirect("/admin/settings/provisioner")
}
```

---

### 6) Acceptance Criteria

1) Settings save and persist after reload.
2) smtp_pass is not cleared when left blank.
3) SMTP test returns ok on success.
4) Changes are logged in audit.

Expected Outcome:
- Settings are manageable via admin UI.

Priority:
HIGH
