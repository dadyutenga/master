Task 10 - Deployment Health Check

Objective:
Add an admin health check endpoint and UI badge for tenant deployments.

---

### 1) Route

Add to main.go:
```go
admin.Get("/tenants/:id/health", h.TenantHealthCheck)
```

---

### 2) Handler

Add TenantHealthCheck in internal/handlers/admin.go:
```go
func (h *Handler) TenantHealthCheck(c *fiber.Ctx) error {
    tenantID := c.Params("id")
    var endpoint string

    err := h.db.QueryRow(
        `SELECT endpoint FROM deployments WHERE tenant_id = ? AND status = 'running' LIMIT 1`,
        tenantID,
    ).Scan(&endpoint)
    if err != nil {
        return c.JSON(fiber.Map{"status": "DOWN", "tenant_id": tenantID, "reason": "no running deployment"})
    }

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(endpoint + "/health")
    status := "DOWN"
    if err == nil && resp.StatusCode == http.StatusOK {
        status = "UP"
        resp.Body.Close()
    }
    return c.JSON(fiber.Map{"status": status, "tenant_id": tenantID})
}
```

Imports needed:
- net/http
- time

---

### 3) UI badge on tenant detail page

In the tenant detail templ file:
```html
<span id="health-badge" class="badge badge-grey">checking...</span>
<script>
fetch('/admin/tenants/{{ .Tenant.ID }}/health')
  .then(r => r.json())
  .then(d => {
    const badge = document.getElementById('health-badge');
    badge.textContent = d.status;
    badge.className = d.status === 'UP' ? 'badge badge-green' : 'badge badge-red';
  })
  .catch(() => {
    document.getElementById('health-badge').textContent = 'ERROR';
  });
</script>
```

---

### 4) Acceptance Criteria

1) Health badge shows UP or DOWN.
2) Endpoint returns DOWN if no running deployment.
3) Errors do not break tenant detail page.

Expected Outcome:
- Admins can see deployment health at a glance.

Priority:
MEDIUM
