Task 9 - Tenant Search, Filter, Pagination

Objective:
Add status filter, search, and pagination to the admin tenant list.

---

### 1) Handler update

Update ListTenants in internal/handlers/admin.go:
```go
func (h *Handler) ListTenants(c *fiber.Ctx) error {
    status := c.Query("status", "")
    search := c.Query("q", "")
    page, _ := strconv.Atoi(c.Query("page", "1"))
    if page < 1 {
        page = 1
    }
    const limit = 20
    offset := (page - 1) * limit

    query := `SELECT id, name, email, status, created_at FROM tenants WHERE 1=1`
    args := []interface{}{}
    if status != "" {
        query += " AND status = ?"
        args = append(args, status)
    }
    if search != "" {
        query += " AND (name LIKE ? OR email LIKE ?)"
        args = append(args, "%"+search+"%", "%"+search+"%")
    }

    countQuery := "SELECT COUNT(*) FROM tenants WHERE 1=1"
    if status != "" {
        countQuery += " AND status = ?"
    }
    if search != "" {
        countQuery += " AND (name LIKE ? OR email LIKE ?)"
    }

    var total int
    h.db.QueryRow(countQuery, args...).Scan(&total)

    query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
    args = append(args, limit, offset)

    rows, err := h.db.Query(query, args...)
    if err != nil {
        return err
    }
    defer rows.Close()

    var tenants []models.Tenant
    for rows.Next() {
        var t models.Tenant
        if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status, &t.CreatedAt); err != nil {
            return err
        }
        tenants = append(tenants, t)
    }

    totalPages := (total + limit - 1) / limit
    return render(c, views.ListTenantsPage(tenants, search, status, page, totalPages))
}
```

---

### 2) View update

Update the tenant list template to add:
- Status filter dropdown
- Search input bound to q
- Pagination controls that preserve q and status

---

### 3) Acceptance Criteria

1) Status filter works.
2) Search matches name or email.
3) Pagination works with correct page count.
4) SQL remains parameterized (no injection).

Expected Outcome:
- Tenant list is searchable and pageable.

Priority:
MEDIUM
