Task 6 - Role Rename (superadmin to admin)

Objective:
Rename the role from superadmin to admin across middleware, routes, and database data so existing superadmins keep access to admin pages.

Stack: Go + Fiber + templ + SQLite
Module: github.com/dadyutenga/hms-control

DO NOT change unrelated auth behavior.

---

### 1) Update role checks in middleware

File: internal/middleware/auth.go

Replace any RequireRole usage:

BEFORE
```go
middleware.RequireRole("superadmin")
```

AFTER
```go
middleware.RequireRole("admin")
```

If RequireRole does a string comparison, ensure it checks for the new role:
```go
func RequireRole(role string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess, err := store.Get(c)
        if err != nil {
            return c.Redirect("/login")
        }
        userRole, ok := sess.Get("role").(string)
        if !ok || userRole != role {
            return c.Status(fiber.StatusForbidden).SendString("Forbidden")
        }
        return c.Next()
    }
}
```

---

### 2) Update admin route group

File: main.go

Update the existing admin group to require "admin":
```go
admin := app.Group("/admin",
    middleware.Auth(store, database),
    middleware.RequireRole("admin"),
)
```

---

### 3) Migration to rename existing roles

Create migrations/005_rename_superadmin.sql:
```sql
UPDATE users SET role = 'admin' WHERE role = 'superadmin';
```

Run migrations:
```sh
go run ./cmd/migrate up
```

---

### 4) Verification

1) Log in with an existing superadmin account.
2) Confirm /admin is accessible (no 403).

Expected Outcome:
- Role rename is complete and backwards compatible.

Priority:
HIGH
