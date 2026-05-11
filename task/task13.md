Task 13 - Tenant Impersonation

Objective:
Allow admins to impersonate a tenant user for troubleshooting, with a clear indicator and safe stop flow.

---

### 1) Routes

Add to main.go:
```go
admin.Post("/tenants/:id/impersonate", h.ImpersonateTenant)
admin.Post("/impersonate/stop",        h.StopImpersonation)
```

---

### 2) Handlers

ImpersonateTenant:
```go
func (h *Handler) ImpersonateTenant(c *fiber.Ctx) error {
    sess, err := h.store.Get(c)
    if err != nil {
        return c.Redirect("/login")
    }
    tenantIDStr := c.Params("id")
    tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
    if err != nil {
        return fiber.ErrBadRequest
    }

    var exists int
    h.db.QueryRow(`SELECT COUNT(*) FROM tenants WHERE id = ?`, tenantID).Scan(&exists)
    if exists == 0 {
        return fiber.ErrNotFound
    }

    adminID := sess.Get("user_id").(int64)
    sess.Set("original_admin_id", adminID)
    sess.Set("impersonating_tenant_id", tenantID)

    LogAction(h.db, adminID, "tenant.impersonated", &tenantID, "", c.IP())

    if err := sess.Save(); err != nil {
        return err
    }
    return c.Redirect("/dashboard")
}
```

StopImpersonation:
```go
func (h *Handler) StopImpersonation(c *fiber.Ctx) error {
    sess, err := h.store.Get(c)
    if err != nil {
        return c.Redirect("/login")
    }

    sess.Delete("impersonating_tenant_id")
    if origID := sess.Get("original_admin_id"); origID != nil {
        sess.Set("user_id", origID)
        sess.Delete("original_admin_id")
    }
    if err := sess.Save(); err != nil {
        return err
    }
    return c.Redirect("/admin/tenants")
}
```

---

### 3) Auth middleware update

In internal/middleware/auth.go, after normal user lookup:
```go
userID := sess.Get("user_id").(int64)

if impersonatedTenantID := sess.Get("impersonating_tenant_id"); impersonatedTenantID != nil {
    var tenantUserID int64
    err := db.QueryRow(
        `SELECT id FROM users WHERE tenant_id = ? LIMIT 1`,
        impersonatedTenantID,
    ).Scan(&tenantUserID)
    if err == nil {
        userID = tenantUserID
    }
}

c.Locals("user_id", userID)
return c.Next()
```

---

### 4) Visible indicator

Add a banner in base layout:
```html
{{ if .ImpersonatingTenantID }}
<div style="background:#f59e0b;color:#000;padding:0.5rem 1rem;text-align:center;">
WARNING: You are impersonating tenant #{{ .ImpersonatingTenantID }}
<form method="POST" action="/impersonate/stop" style="display:inline;margin-left:1rem;">
<button type="submit">Stop Impersonation</button>
</form>
</div>
{{ end }}
```

---

### 5) Acceptance Criteria

1) Admin can impersonate a tenant.
2) Tenant pages render as the tenant user.
3) Banner shows and stop restores admin session.
4) Audit log records tenant.impersonated.

Expected Outcome:
- Safe admin impersonation with visibility.

Priority:
HIGH
