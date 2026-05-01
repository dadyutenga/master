package handlers

import (
	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/views/admin"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
	return c.Redirect("/admin/tenants")
}

func (h *Handler) ListTenants(c *fiber.Ctx) error {
	q := generated.New(h.db)
	tenants, err := q.ListTenants(c.Context(), generated.ListTenantsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	if c.Get("HX-Request") == "true" {
		return render(c, admin.TenantTable(tenants))
	}
	return render(c, admin.TenantList(tenants))
}

func (h *Handler) ShowTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	user, err := q.GetUserByID(c.Context(), tenant.UserID)
	if err != nil {
		return fiber.ErrNotFound
	}

	return render(c, admin.TenantDetail(tenant, user))
}

func (h *Handler) ApproveTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}
	if tenant.Status != generated.TenantStatusPendingApproval {
		return c.Status(409).SendString("Tenant not in pending_approval state")
	}

	q.ApproveTenant(c.Context(), id)

	h.eng.Enqueue(id)

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}
	return c.Redirect("/admin/tenants")
}

func (h *Handler) SuspendTenant(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	q := generated.New(h.db)
	q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
		ID:     id,
		Status: generated.TenantStatusSuspended,
	})

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusSuspended))
	}
	return c.Redirect("/admin/tenants")
}

func (h *Handler) RetryProvision(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	q := generated.New(h.db)

	tenant, _ := q.GetTenantByID(c.Context(), id)
	if tenant.Status != generated.TenantStatusFailed {
		return c.Status(409).SendString("Only failed tenants can be retried")
	}

	q.SetTenantProvisioning(c.Context(), id)
	h.eng.Enqueue(id)

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}
	return c.Redirect("/admin/tenants/" + id.String())
}