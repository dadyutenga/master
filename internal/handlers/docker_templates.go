package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/models"
	"github.com/dadyutenga/hms-control/internal/views/admin"

	"github.com/gofiber/fiber/v2"
)

// GET /admin/docker-templates
func (h *Handler) ListDockerTemplates(c *fiber.Ctx) error {
	store := models.NewDockerTemplateStore(h.db)
	templates, err := store.List()
	if err != nil {
		return fmt.Errorf("list docker templates: %w", err)
	}
	return render(c, admin.DockerTemplatesList(templates))
}

// GET /admin/docker-templates/new
func (h *Handler) ShowCreateDockerTemplate(c *fiber.Ctx) error {
	return render(c, admin.DockerTemplateForm(nil, "", ""))
}

// POST /admin/docker-templates/new
func (h *Handler) CreateDockerTemplate(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))
	body := strings.TrimSpace(c.FormValue("template_body"))

	if name == "" {
		return render(c, admin.DockerTemplateForm(nil, "Template name is required.", ""))
	}
	if body == "" {
		return render(c, admin.DockerTemplateForm(nil, "Template body is required.", name))
	}

	store := models.NewDockerTemplateStore(h.db)
	if store.NameExists(name, 0) {
		return render(c, admin.DockerTemplateForm(nil, "A template with that name already exists.", name))
	}

	tmpl, err := store.Create(name, description, body)
	if err != nil {
		return render(c, admin.DockerTemplateForm(nil, "Failed to create template: "+err.Error(), name))
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "docker_template.created", nil, fmt.Sprintf("created template '%s'", tmpl.Name), c.IP())
	}

	return c.Redirect("/admin/docker-templates")
}

// GET /admin/docker-templates/:id/edit
func (h *Handler) ShowEditDockerTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid template ID")
	}

	store := models.NewDockerTemplateStore(h.db)
	tmpl, err := store.GetByID(id)
	if err != nil {
		return fiber.ErrNotFound
	}

	return render(c, admin.DockerTemplateForm(tmpl, "", ""))
}

// POST /admin/docker-templates/:id/edit
func (h *Handler) UpdateDockerTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid template ID")
	}

	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))
	body := strings.TrimSpace(c.FormValue("template_body"))

	if name == "" {
		return render(c, admin.DockerTemplateForm(nil, "Template name is required.", ""))
	}
	if body == "" {
		return render(c, admin.DockerTemplateForm(nil, "Template body is required.", name))
	}

	store := models.NewDockerTemplateStore(h.db)
	tmpl, _ := store.GetByID(id)
	if tmpl == nil {
		return fiber.ErrNotFound
	}

	if store.NameExists(name, id) {
		return render(c, admin.DockerTemplateForm(tmpl, "A template with that name already exists.", name))
	}

	if err := store.Update(id, name, description, body); err != nil {
		return render(c, admin.DockerTemplateForm(tmpl, "Failed to update template: "+err.Error(), name))
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "docker_template.updated", nil, fmt.Sprintf("updated template '%s'", name), c.IP())
	}

	return c.Redirect("/admin/docker-templates")
}

// POST /admin/docker-templates/:id/delete
func (h *Handler) DeleteDockerTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid template ID")
	}

	store := models.NewDockerTemplateStore(h.db)
	tmpl, _ := store.GetByID(id)
	if tmpl == nil {
		return fiber.ErrNotFound
	}

	if err := store.Delete(id); err != nil {
		return c.Status(500).SendString("Failed to delete template.")
	}

	if uid, ok := middleware.GetUserID(c); ok {
		h.audit.Log(uid, "docker_template.deleted", nil, fmt.Sprintf("deleted template '%s'", tmpl.Name), c.IP())
	}

	return c.Redirect("/admin/docker-templates")
}
