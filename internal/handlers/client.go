package handlers

import (
	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/models"
	"github.com/dadyutenga/hms-control/internal/views/client"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ClientDashboard(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}

	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	user, err := q.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(500).SendString("Failed to load user.")
	}

	contact, err := h.contactDetails(c)
	if err != nil {
		return err
	}

	if tenant.Status == generated.TenantStatusPendingVerification || tenant.Status == generated.TenantStatusPendingApproval {
		return render(c, client.Pending(client.PendingProps{
			Tenant:  tenant,
			User:    user,
			Contact: contact,
		}))
	}

	// Pause container if tenant is active but unpaid
	if tenant.Status == generated.TenantStatusActive && tenant.BillingStatus != generated.BillingStatusPaid {
		h.pauseTenantContainer(tenant.ID)
	}

	// Get instance stats
	var stats models.ClientDashboardStats
	instances, err := q.ListInstancesByTenantID(c.Context(), tenant.ID)
	if err == nil {
		stats.TotalInstances = len(instances)
		for _, inst := range instances {
			switch inst.Status {
			case "active":
				stats.ActiveInstances++
			case "paused":
				stats.PausedInstances++
			case "disabled":
				stats.DisabledInstances++
			case "failed":
				stats.FailedInstances++
			}
		}
	}

	return render(c, client.Dashboard(client.DashboardProps{
		Tenant: tenant,
		User:   user,
		Stats:  stats,
	}))
}

func (h *Handler) pauseTenantContainer(tenantID interface{}) {
	// Best-effort pause - don't fail if container doesn't exist
	id, ok := tenantID.(interface{ String() string })
	if !ok {
		return
	}
	h.db.Exec(`UPDATE deployments SET status = 'paused', updated_at = CURRENT_TIMESTAMP WHERE tenant_id = ? AND status = 'active'`, id.String())
}
