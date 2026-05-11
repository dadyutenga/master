package middleware

import (
	"database/sql"

	"github.com/dadyutenga/hms-control/internal/db/generated"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func Auth(store *session.Store, db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		rawID := sess.Get("userID")
		if rawID == nil {
			return c.Redirect("/login")
		}
		userID, ok := rawID.(int64)
		if !ok {
			sess.Destroy()
			return c.Redirect("/login")
		}

		// Check for impersonation — load impersonated tenant user
		if impIDRaw := sess.Get("impersonating_tenant_id"); impIDRaw != nil {
			impID, _ := impIDRaw.(string)
			if impID != "" {
				var user generated.User
				err := db.QueryRow(
					`SELECT u.id, u.name, u.email, u.company, u.phone, u.password,
					        u.role, u.verified, u.tin, u.brela_number, u.created_at, u.updated_at
					 FROM users u
					 JOIN tenants t ON t.user_id = u.id
					 WHERE t.id = ? AND u.role = 'client'`, impID,
				).Scan(&user.ID, &user.Name, &user.Email, &user.Company,
					&user.Phone, &user.Password, &user.Role, &user.Verified,
					&user.TIN, &user.BrelaNumber, &user.CreatedAt, &user.UpdatedAt)
				if err == nil {
					c.Locals("user", user)
					c.Locals("user_role", user.Role)
					c.Locals("impersonating", true)
					c.Locals("original_admin_id", userID)
					return c.Next()
				}
				// Target gone — clean up
				sess.Delete("impersonating_tenant_id")
				sess.Save()
			}
		}

		q := generated.New(db)
		user, err := q.GetUserByID(c.Context(), userID)
		if err != nil {
			sess.Destroy()
			return c.Redirect("/login")
		}

		c.Locals("user", user)
		c.Locals("user_role", user.Role)
		c.Locals("impersonating", false)
		return c.Next()
	}
}

func GetUser(c *fiber.Ctx) (generated.User, bool) {
	u, ok := c.Locals("user").(generated.User)
	return u, ok
}

func GetUserID(c *fiber.Ctx) (int64, bool) {
	u, ok := GetUser(c)
	if !ok {
		return 0, false
	}
	return u.ID, true
}

func IsImpersonating(c *fiber.Ctx) bool {
	imp, _ := c.Locals("impersonating").(bool)
	return imp
}
