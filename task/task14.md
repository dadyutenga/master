Task 14 - Export Tenants CSV

Objective:
Provide CSV export for tenant list and ensure route order is correct.

---

### 1) Route order (critical)

In main.go, register /tenants/export BEFORE /tenants/:id:
```go
admin.Get("/tenants/export", h.ExportTenantsCSV)
admin.Get("/tenants/:id",    h.ShowTenant)
```

---

### 2) ExportTenantsCSV handler

Add to internal/handlers/admin.go:
```go
func (h *Handler) ExportTenantsCSV(c *fiber.Ctx) error {
    c.Set("Content-Type",        "text/csv")
    c.Set("Content-Disposition", `attachment; filename="tenants.csv"`)

    rows, err := h.db.Query(
        `SELECT id, name, email, status, created_at FROM tenants ORDER BY created_at DESC`,
    )
    if err != nil {
        return err
    }
    defer rows.Close()

    w := csv.NewWriter(c.Response().BodyWriter())
    if err := w.Write([]string{"ID", "Name", "Email", "Status", "Created At"}); err != nil {
        return err
    }

    for rows.Next() {
        var (
            id                          int64
            name, email, status, createdAt string
        )
        if err := rows.Scan(&id, &name, &email, &status, &createdAt); err != nil {
            return err
        }
        if err := w.Write([]string{
            strconv.FormatInt(id, 10), name, email, status, createdAt,
        }); err != nil {
            return err
        }
    }

    w.Flush()
    return w.Error()
}
```

Imports needed:
- encoding/csv
- strconv

---

### 3) Acceptance Criteria

1) CSV downloads with correct headers.
2) Export works with many tenants.
3) /tenants/export does not hit ShowTenant.

Expected Outcome:
- Admin can export tenants to CSV.

Priority:
MEDIUM
