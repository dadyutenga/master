package handlers

import (
	"strings"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/views/client"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) ClientInstancesList(c *fiber.Ctx) error {
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
		return c.Status(500).SendString("Failed to load user")
	}

	instances, err := q.ListInstancesByTenantID(c.Context(), tenant.ID)
	if err != nil {
		instances = []generated.Instance{}
	}

	return render(c, client.InstancesList(client.InstancesListProps{
		User:      user,
		Instances: instances,
	}))
}

func (h *Handler) ClientInstanceDetail(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}
	user, err := q.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(500).SendString("Failed to load user")
	}

	inst, err := q.GetInstanceByID(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}
	if inst.TenantID != tenant.ID {
		return c.Status(403).SendString("Access denied")
	}

	deployments, err := q.ListInstanceDeployments(c.Context(), id)
	if err != nil {
		deployments = []generated.InstanceDeployment{}
	}

	return render(c, client.InstanceDetail(client.InstanceDetailProps{
		User:        user,
		Instance:    inst,
		Deployments: deployments,
	}))
}

func (h *Handler) ClientPauseInstance(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	inst, err := q.GetInstanceByID(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}
	if inst.TenantID != tenant.ID {
		return c.Status(403).SendString("Access denied")
	}
	if inst.Status != "active" {
		return c.Status(409).SendString("Only active instances can be paused")
	}

	if err := q.UpdateInstanceStatus(c.Context(), id, "paused"); err != nil {
		return c.Status(500).SendString("Failed to pause instance")
	}

	q.CreateInstanceDeployment(c.Context(), id, "pause", "active")

	if c.Get("HX-Request") == "true" {
		return c.Redirect("/dashboard/instances/" + id.String())
	}
	return c.Redirect("/dashboard/instances/" + id.String())
}

func (h *Handler) ClientRestartInstance(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	inst, err := q.GetInstanceByID(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}
	if inst.TenantID != tenant.ID {
		return c.Status(403).SendString("Access denied")
	}
	if inst.Status != "paused" {
		return c.Status(409).SendString("Only paused instances can be restarted")
	}

	if err := q.UpdateInstanceStatus(c.Context(), id, "active"); err != nil {
		return c.Status(500).SendString("Failed to restart instance")
	}

	q.CreateInstanceDeployment(c.Context(), id, "restart", "active")

	if c.Get("HX-Request") == "true" {
		return c.Redirect("/dashboard/instances/" + id.String())
	}
	return c.Redirect("/dashboard/instances/" + id.String())
}

func (h *Handler) ClientDisableInstance(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}
	userID, ok := sess.Get("userID").(int64)
	if !ok {
		return c.Redirect("/login")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	inst, err := q.GetInstanceByID(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}
	if inst.TenantID != tenant.ID {
		return c.Status(403).SendString("Access denied")
	}
	if inst.Status != "active" && inst.Status != "paused" {
		return c.Status(409).SendString("Instance cannot be disabled in current state")
	}

	if err := q.SetInstanceAdminDisabled(c.Context(), id, true); err != nil {
		return c.Status(500).SendString("Failed to disable instance")
	}

	q.CreateInstanceDeployment(c.Context(), id, "disable", "active")

	if c.Get("HX-Request") == "true" {
		return c.Redirect("/dashboard/instances/" + id.String())
	}
	return c.Redirect("/dashboard/instances/" + id.String())
}

func (h *Handler) ClientShownewInstance(c *fiber.Ctx) error {
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
		return c.Status(500).SendString("Failed to load user")
	}

	packages, err := q.ListBillingPackages(c.Context())
	if err != nil {
		packages = []generated.BillingPackage{}
	}

	return render(c, client.NewInstance(client.NewInstanceProps{
		User:     user,
		Tenant:   tenant,
		Packages: packages,
	}))
}

func (h *Handler) ClientCreateInstance(c *fiber.Ctx) error {
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

	hotelName := c.FormValue("hotel_name")
	slug := strings.ToLower(strings.ReplaceAll(c.FormValue("slug"), " ", "-"))
	pkgName := c.FormValue("package")

	if hotelName == "" || slug == "" {
		return c.Status(400).SendString("Hotel name and slug are required")
	}

	// Find package price
	packages, _ := q.ListBillingPackages(c.Context())
	var price float64
	for _, p := range packages {
		if p.Name == pkgName {
			price = p.Price
			break
		}
	}

	dbName := "hms_" + slug
	dbUser := "hms_" + slug
	dbPass := "pass_" + slug

	inst, err := q.CreateInstance(c.Context(), tenant.ID, hotelName, slug, dbName, dbUser, dbPass, pkgName, price)
	if err != nil {
		return c.Status(500).SendString("Failed to create instance: " + err.Error())
	}

	return c.Redirect("/dashboard/instances/" + inst.ID.String())
}
