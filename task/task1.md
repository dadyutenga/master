Task: Admin Contact Details From Database

Objective:
Allow admins to manage and update contact details (location and phone number) with values fetched from the database instead of hardcoded values.

This task must:
- Replace any hardcoded contact details in admin-facing views or templates
- Add a persisted source of truth in the database for contact details
- Provide admin UI and endpoints to update contact details
- Reuse existing validation, handlers, and view patterns

DO NOT hardcode contact details in templates or handlers.
DO NOT bypass server-side validation.
DO NOT change unrelated flows.

---

### 1) Discovery First: Find Existing Patterns

Before implementation, identify:
- Where contact details are currently rendered (templates, handlers)
- Existing admin settings or configuration storage patterns
- Validation and error handling conventions
- Database access helpers and query structure

Implementation rule:
- Extend existing repository patterns and query helpers
- Avoid introducing new libraries

---

### 2) Database Model and Queries

Required behavior:
- Add a table or settings record for contact details if one does not exist
- Fields must include: location, phone_number
- Provide queries to fetch and update these values

Implementation notes:
- Use existing migration and query generator patterns
- Include created_at and updated_at if that matches existing schemas

---

### 3) Admin UI and Endpoints

Required behavior:
- Admin UI can view current contact details
- Admin UI can update location and phone number
- On save, persist and re-render updated values

Implementation notes:
- Reuse admin layout and forms if possible
- Show inline validation errors

---

### 4) Rendering in Views

Required behavior:
- All admin contact displays must read from database values
- No hardcoded strings for location or phone number

---

### 5) Validation and Security

Required behavior:
- Validate required fields server-side
- Normalize phone number format if existing utilities are used
- Only admins can update contact details

---

### 6) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Contact details render from database on admin pages
2) Admin can update location and phone number
3) Values persist across reloads
4) Non-admins cannot update contact details

---

Expected Outcome:
- Admin contact details are fully data-driven from the database

Priority:
HIGH
