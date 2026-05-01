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

		userID := sess.Get("userID")
		if userID == nil {
			return c.Redirect("/login")
		}

		q := generated.New(db)
		user, err := q.GetUserByID(c.Context(), userID.(int64))
		if err != nil {
			sess.Destroy()
			return c.Redirect("/login")
		}

		c.Locals("user", user)
		return c.Next()
	}
}

func GetUser(c *fiber.Ctx) generated.User {
	return c.Locals("user").(generated.User)
}