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

	deployments, err := q.ListDeploymentsByTenantID(c.Context(), tenant.ID)
	if err != nil {
		return err
	}

	return render(c, admin.TenantDetail(tenant, user, deployments))
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

	if err := q.ApproveTenant(c.Context(), id); err != nil {
		return c.Status(500).SendString("Failed to approve tenant.")
	}

	if _, err := q.CreateDeployment(c.Context(), generated.CreateDeploymentParams{
		TenantID: id,
		Action:   "approve",
		Status:   generated.DeploymentStatusProvisioning,
	}); err != nil {
		return c.Status(500).SendString("Failed to record deployment.")
	}

	h.eng.Enqueue(id)

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}
	return c.Redirect("/admin/tenants")
}

func (h *Handler) SuspendTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}
	if tenant.Status != generated.TenantStatusActive {
		return c.Status(409).SendString("Only active tenants can be suspended")
	}

	if err := q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
		ID:     id,
		Status: generated.TenantStatusSuspended,
	}); err != nil {
		return c.Status(500).SendString("Failed to suspend tenant.")
	}

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusSuspended))
	}
	return c.Redirect("/admin/tenants")
}

func (h *Handler) RetryProvision(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid tenant ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}
	if tenant.Status != generated.TenantStatusFailed {
		return c.Status(409).SendString("Only failed tenants can be retried")
	}

	if err := q.SetTenantProvisioning(c.Context(), id); err != nil {
		return c.Status(500).SendString("Failed to reset tenant status.")
	}
	if _, err := q.CreateDeployment(c.Context(), generated.CreateDeploymentParams{
		TenantID: id,
		Action:   "retry",
		Status:   generated.DeploymentStatusProvisioning,
	}); err != nil {
		return c.Status(500).SendString("Failed to record deployment.")
	}
	h.eng.Enqueue(id)

	if c.Get("HX-Request") == "true" {
		return render(c, admin.StatusBadge(generated.TenantStatusProvisioning))
	}
	return c.Redirect("/admin/tenants/" + id.String())
}
