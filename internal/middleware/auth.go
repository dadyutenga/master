package middleware

import (
	"database/sql"

	"github.com/dadyutenga/hms-control/internal/db/generated"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
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
		if !ok || userID == 0 {
			sess.Destroy()
			return c.Redirect("/login")
		}

		q := generated.New(db)

		if impRaw := sess.Get("impersonating_tenant_id"); impRaw != nil {
			if impID, ok := impRaw.(string); ok && impID != "" {
				tenantID, err := uuid.Parse(impID)
				if err != nil {
					sess.Delete("impersonating_tenant_id")
					sess.Save()
					return c.Redirect("/admin/tenants")
				}
				tenant, err := q.GetTenantByID(c.Context(), tenantID)
				if err == nil {
					user, err := q.GetUserByID(c.Context(), tenant.UserID)
					if err == nil {
						c.Locals("user", user)
						c.Locals("user_id", user.ID)
						c.Locals("user_email", user.Email)
						c.Locals("user_role", user.Role)
						c.Locals("impersonating", true)
						if adminID, ok := sess.Get("original_admin_id").(int64); ok {
							c.Locals("original_admin_id", adminID)
						}
						return c.Next()
					}
				}
				sess.Delete("impersonating_tenant_id")
				sess.Save()
				return c.Redirect("/admin/tenants")
			}
		}

		user, err := q.GetUserByID(c.Context(), userID)
		if err != nil {
			sess.Destroy()
			return c.Redirect("/login")
		}

		c.Locals("user", user)
		c.Locals("user_id", user.ID)
		c.Locals("user_email", user.Email)
		c.Locals("user_role", user.Role)
		c.Locals("impersonating", false)
		return c.Next()
	}
}

func GetUser(c *fiber.Ctx) (generated.User, bool) {
	u, ok := c.Locals("user").(generated.User)
	return u, ok
}
