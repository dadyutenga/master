package handlers

import (
	"regexp"
	"strings"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/views/client"

	"github.com/gofiber/fiber/v2"
)

var subdomainPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func (h *Handler) ShowTenantDetails(c *fiber.Ctx) error {
	user, ok := middleware.GetUser(c)
	if !ok {
		return c.Redirect("/login")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), user.ID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	return render(c, client.TenantDetails(client.TenantDetailsProps{
		Tenant: tenant,
		User:   user,
	}))
}

func (h *Handler) UpdateTenantDetails(c *fiber.Ctx) error {
	user, ok := middleware.GetUser(c)
	if !ok {
		return c.Redirect("/login")
	}

	q := generated.New(h.db)
	tenant, err := q.GetTenantByUserID(c.Context(), user.ID)
	if err != nil {
		return c.Status(404).SendString("No tenant found")
	}

	subdomain := strings.ToLower(strings.TrimSpace(c.FormValue("requested_subdomain")))
	adminName := strings.TrimSpace(c.FormValue("admin_name"))
	adminEmail := strings.TrimSpace(c.FormValue("admin_email"))
	adminPhone := strings.TrimSpace(c.FormValue("admin_phone"))

	if subdomain == "" || adminName == "" || adminEmail == "" {
		return render(c, client.TenantDetails(client.TenantDetailsProps{
			Tenant: tenant,
			User:   user,
			Error:  "Subdomain, admin name, and admin email are required.",
		}))
	}
	if !subdomainPattern.MatchString(subdomain) {
		return render(c, client.TenantDetails(client.TenantDetailsProps{
			Tenant: tenant,
			User:   user,
			Error:  "Subdomain must be lowercase and use only letters, numbers, and hyphens.",
		}))
	}

	if existing, err := q.GetTenantByRequestedSubdomain(c.Context(), subdomain); err == nil && existing.ID != tenant.ID {
		return render(c, client.TenantDetails(client.TenantDetailsProps{
			Tenant: tenant,
			User:   user,
			Error:  "That subdomain is already reserved.",
		}))
	}
	if existing, err := q.GetTenantBySlug(c.Context(), subdomain); err == nil && existing.ID != tenant.ID {
		return render(c, client.TenantDetails(client.TenantDetailsProps{
			Tenant: tenant,
			User:   user,
			Error:  "That subdomain is already in use.",
		}))
	}

	if err := q.UpdateTenantDetails(c.Context(), generated.UpdateTenantDetailsParams{
		ID:                 tenant.ID,
		RequestedSubdomain: &subdomain,
		AdminName:          &adminName,
		AdminEmail:         &adminEmail,
		AdminPhone:         nullString(adminPhone),
	}); err != nil {
		return render(c, client.TenantDetails(client.TenantDetailsProps{
			Tenant: tenant,
			User:   user,
			Error:  "Failed to update tenant details.",
		}))
	}

	tenant, err = q.GetTenantByID(c.Context(), tenant.ID)
	if err != nil {
		return err
	}

	return render(c, client.TenantDetails(client.TenantDetailsProps{
		Tenant:  tenant,
		User:    user,
		Success: true,
	}))
}
