package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := GetUser(c)
		if !ok {
			return c.Status(401).SendString("Unauthorized")
		}
		if user.Role != role {
			return c.Status(403).SendString("Forbidden")
		}
		return c.Next()
	}
}