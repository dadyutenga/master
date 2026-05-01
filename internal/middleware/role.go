package middleware

import (
	"github.com/dadyutenga/hms-control/internal/db/generated"

	"github.com/gofiber/fiber/v2"
)

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(generated.User)
		if user.Role != role {
			return c.Status(403).SendString("Forbidden")
		}
		return c.Next()
	}
}