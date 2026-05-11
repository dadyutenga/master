Task: Tenant Details Entry Views and Fields

Objective:
Create tenant-facing views and entries for tenant users to input required details, including the subdomain to be used and the admin account to be used with the system. For now, fields only (no provisioning or automation).

This task must:
- Add tenant-facing form views to capture details
- Persist tenant inputs using existing handlers and validation patterns
- Include fields for subdomain and admin account details
- Reuse existing layouts and templating conventions

DO NOT auto-provision or run scripts.
DO NOT hardcode tenant values.
DO NOT change unrelated flows.

---

### 1) Discovery First: Find Existing Tenant Patterns

Before implementation, identify:
- Tenant user model and any existing tenant profile fields
- Current tenant dashboard or settings pages
- Existing form handling and validation patterns

Implementation rule:
- Extend existing view and handler patterns
- Avoid new dependencies

---

### 2) Required Fields (Fields Only)

Required behavior:
- Subdomain to be used (required, validated for format)
- Admin account details to be used with the system (required fields only)

Admin account field suggestions (confirm with existing models):
- Admin full name
- Admin email
- Admin phone (optional if pattern exists)

---

### 3) Views and Handlers

Required behavior:
- Tenant user can view and edit their details
- Form shows current values on load
- On submit, validate and persist
- Show inline validation errors

Implementation notes:
- Use existing templ layout for tenant pages
- Ensure correct tenant scoping

---

### 4) Validation and Security

Required behavior:
- Server-side validation for required fields
- Subdomain validation: lowercase, alphanumeric and hyphen, no spaces
- Enforce uniqueness if pattern exists
- Prevent cross-tenant updates

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Tenant can view and edit subdomain and admin account fields
2) Invalid subdomain formats are rejected
3) Updates persist and reload correctly
4) Users cannot edit other tenants' data

---

Expected Outcome:
- Tenant details entry views exist and persist fields only

Priority:
HIGH
