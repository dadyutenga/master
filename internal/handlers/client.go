package handlers

import (
	"github.com/dadyutenga/hms-control/internal/db/generated"
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

	return render(c, client.Dashboard(client.DashboardProps{
		Tenant: tenant,
		User:   user,
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
