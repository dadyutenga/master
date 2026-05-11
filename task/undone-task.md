Undone Task: Custom Docker Templates for Admin

Objective:
Add admin-facing views and backend functionality so admins can create and manage custom Docker templates for deployments.

This task must:
- Provide an admin UI to add, edit, and delete Docker templates
- Persist templates in the database
- Allow provisioner settings to select from custom templates
- Reuse existing admin layout and form patterns

DO NOT hardcode templates in code.
DO NOT expose template management to non-admins.

---

### 1) Discovery First: Find Existing Provisioner Settings

Before implementation, identify:
- Current provisioner settings pages and handlers
- Where docker_template is stored and used
- Existing admin UI patterns for settings forms

---

### 2) Data Model

Required behavior:
- Create a docker_templates table with fields:
  - id (primary key)
  - name (unique, human-friendly)
  - template_path or template_body (depending on current template storage)
  - created_at, updated_at

---

### 3) Admin UI

Required behavior:
- Admin list page showing templates
- Create form with validation
- Edit and delete actions with confirmation

---

### 4) Integration with Provisioner Settings

Required behavior:
- Provisioner settings page should list available templates
- Selected template is stored in settings and used during deployment

---

### 5) Validation and Security

Required behavior:
- Validate template name format and uniqueness
- Admin-only access for all template endpoints
- Sanitize or validate template content/path

---

### 6) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Admin can add a new template
2) Admin can edit and delete a template
3) Provisioner settings can select a custom template
4) Selected template is used during deployments

---

Expected Outcome:
- Admins can manage custom Docker templates for deployments

Priority:
HIGH
