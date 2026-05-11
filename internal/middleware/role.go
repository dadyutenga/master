package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Block access during impersonation — admin is acting as a tenant
		if imp, _ := c.Locals("impersonating").(bool); imp {
			return fiber.ErrForbidden
		}
		r, _ := c.Locals("user_role").(string)
		if r != role {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}
