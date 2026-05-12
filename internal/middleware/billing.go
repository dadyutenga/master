package middleware

import (
	"database/sql"

	"github.com/dadyutenga/hms-control/internal/db/generated"

	"github.com/gofiber/fiber/v2"
)

func RequireBilling(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := GetUser(c)
		if !ok {
			return c.Status(401).SendString("Unauthorized")
		}

		q := generated.New(db)
		tenant, err := q.GetTenantByUserID(c.Context(), user.ID)
		if err != nil {
			return c.Status(404).SendString("No tenant found")
		}

		c.Locals("tenant", tenant)
		return c.Next()
	}
}

func GetTenant(c *fiber.Ctx) (generated.Tenant, bool) {
	t, ok := c.Locals("tenant").(generated.Tenant)
	return t, ok
}
