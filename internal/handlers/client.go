package handlers

import (
	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/views/client"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ClientDashboard(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	userID := sess.Get("userID").(int64)

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	user, _ := q.GetUserByID(c.Context(), userID)

	return render(c, client.Dashboard(client.DashboardProps{
		Tenant: tenant,
		User:   user,
	}))
}