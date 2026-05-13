package handlers

import (
	"strconv"
	"time"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/views/admin"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) AdminInstancesList(c *fiber.Ctx) error {
	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	q := generated.New(h.db)
	instances, err := q.ListAllInstances(c.Context())
	if err != nil {
		instances = []generated.Instance{}
	}

	// Filter by status
	var filtered []generated.Instance
	if status != "" {
		for _, inst := range instances {
			if inst.Status == status {
				filtered = append(filtered, inst)
			}
		}
	} else {
		filtered = instances
	}

	const limit = 20
	total := len(filtered)
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	paged := filtered[start:end]

	return render(c, admin.AdminInstancesList(admin.AdminInstancesListProps{
		Instances:  paged,
		Status:     status,
		Page:       page,
		TotalPages: totalPages,
	}))
}

func (h *Handler) AdminInstanceDetail(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	inst, err := q.GetInstanceByID(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}

	deployments, err := q.ListInstanceDeployments(c.Context(), id)
	if err != nil {
		deployments = []generated.InstanceDeployment{}
	}

	return render(c, admin.AdminInstanceDetail(admin.AdminInstanceDetailProps{
		Instance:    inst,
		Deployments: deployments,
	}))
}

func (h *Handler) AdminEnableInstance(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	if err := q.SetInstanceAdminDisabled(c.Context(), id, false); err != nil {
		return c.Status(500).SendString("Failed to enable instance")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "instance.enabled", &tid, "", c.IP())
	}

	return c.Redirect("/admin/instances/" + id.String())
}

func (h *Handler) AdminDisableInstance(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	if err := q.SetInstanceAdminDisabled(c.Context(), id, true); err != nil {
		return c.Status(500).SendString("Failed to disable instance")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "instance.disabled", &tid, "", c.IP())
	}

	return c.Redirect("/admin/instances/" + id.String())
}

func (h *Handler) AdminStartInstance(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	if err := q.UpdateInstanceStatus(c.Context(), id, "active"); err != nil {
		return c.Status(500).SendString("Failed to start instance")
	}

	q.CreateInstanceDeployment(c.Context(), id, "start", "active")

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "instance.started", &tid, "", c.IP())
	}

	return c.Redirect("/admin/instances/" + id.String())
}

func (h *Handler) AdminStopInstance(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	if err := q.UpdateInstanceStatus(c.Context(), id, "paused"); err != nil {
		return c.Status(500).SendString("Failed to stop instance")
	}

	q.CreateInstanceDeployment(c.Context(), id, "stop", "active")

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "instance.stopped", &tid, "", c.IP())
	}

	return c.Redirect("/admin/instances/" + id.String())
}

func (h *Handler) AdminArchiveInstance(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	if err := q.ArchiveInstance(c.Context(), id); err != nil {
		return c.Status(500).SendString("Failed to archive instance")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "instance.archived", &tid, "", c.IP())
	}

	return c.Redirect("/admin/instances")
}

func (h *Handler) AdminDeleteInstance(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	q := generated.New(h.db)
	if err := q.DeleteInstance(c.Context(), id); err != nil {
		return c.Status(500).SendString("Failed to delete instance")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "instance.deleted", &tid, "", c.IP())
	}

	return c.Redirect("/admin/instances")
}

func (h *Handler) AdminUpdateInstancePrice(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	priceStr := c.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid price")
	}

	packageName := c.FormValue("package_name")
	if packageName == "" {
		packageName = "basic"
	}

	q := generated.New(h.db)
	if err := q.UpdateInstancePrice(c.Context(), id, price, packageName); err != nil {
		return c.Status(500).SendString("Failed to update price")
	}

	return c.Redirect("/admin/instances/" + id.String())
}

func (h *Handler) AdminUpdateInstanceBilling(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).SendString("Invalid instance ID")
	}

	status := c.FormValue("billing_status")
	if status != "unpaid" && status != "paid" && status != "overdue" && status != "suspended" {
		return c.Status(400).SendString("Invalid billing status")
	}

	q := generated.New(h.db)

	var lastPayment, nextDue *time.Time
	if lp := c.FormValue("last_payment_at"); lp != "" {
		t, err := time.Parse("2006-01-02", lp)
		if err == nil {
			lastPayment = &t
		}
	}
	if nd := c.FormValue("next_due_at"); nd != "" {
		t, err := time.Parse("2006-01-02", nd)
		if err == nil {
			nextDue = &t
		}
	}

	if err := q.UpdateInstanceBilling(c.Context(), id, status, lastPayment, nextDue); err != nil {
		return c.Status(500).SendString("Failed to update instance billing")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		tid := id.String()
		h.audit.Log(uid, "instance.billing.updated", &tid, "billing status -> "+status, c.IP())
	}

	return c.Redirect("/admin/instances/" + id.String())
}

// Billing Packages

func (h *Handler) AdminBillingPackages(c *fiber.Ctx) error {
	q := generated.New(h.db)
	packages, err := q.GetAllBillingPackages(c.Context())
	if err != nil {
		packages = []generated.BillingPackage{}
	}

	return render(c, admin.BillingPackagesList(packages))
}

func (h *Handler) AdminCreateBillingPackage(c *fiber.Ctx) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	priceStr := c.FormValue("price")
	billingCycle := c.FormValue("billing_cycle")

	if name == "" {
		return c.Status(400).SendString("Name is required")
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid price")
	}

	q := generated.New(h.db)
	if _, err := q.CreateBillingPackage(c.Context(), name, description, price, "TZS", billingCycle); err != nil {
		return c.Status(500).SendString("Failed to create package")
	}

	return c.Redirect("/admin/billing-packages")
}

func (h *Handler) AdminEditBillingPackage(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid package ID")
	}

	q := generated.New(h.db)
	pkg, err := q.GetBillingPackageByID(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	return render(c, admin.BillingPackageEdit(pkg))
}

func (h *Handler) AdminUpdateBillingPackage(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid package ID")
	}

	name := c.FormValue("name")
	description := c.FormValue("description")
	priceStr := c.FormValue("price")
	billingCycle := c.FormValue("billing_cycle")
	isActive := c.FormValue("is_active") == "1"

	if name == "" {
		return c.Status(400).SendString("Name is required")
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid price")
	}

	q := generated.New(h.db)
	if err := q.UpdateBillingPackage(c.Context(), id, name, description, price, billingCycle, isActive); err != nil {
		return c.Status(500).SendString("Failed to update package")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "billing_package.updated", nil, "updated package '"+name+"'", c.IP())
	}

	return c.Redirect("/admin/billing-packages")
}

func (h *Handler) AdminDeleteBillingPackage(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid package ID")
	}

	q := generated.New(h.db)
	if err := q.DeleteBillingPackage(c.Context(), id); err != nil {
		return c.Status(500).SendString("Failed to delete package")
	}

	return c.Redirect("/admin/billing-packages")
}
