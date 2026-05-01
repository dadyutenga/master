Task: Superadmin Accounts and Management Templates

Objective:
Create three secure superadmin accounts and a management interface for superadmin administration, including access control and audit visibility.

This task must:
- Seed three superadmin accounts securely
- Enforce superadmin-only access to admin templates and endpoints
- Provide management screens for superadmin actions
- Reuse existing auth and role patterns

DO NOT store plaintext passwords.
DO NOT expose admin pages to non-admin roles.
DO NOT add new dependencies without approval.

---

### 1) Discovery First: Find Existing Auth and Role Patterns

Before implementation, identify:
- Current user roles and permissions
- Auth middleware and session handling
- Any existing admin account creation or seeding
- Audit logging patterns

Implementation rule:
- Extend existing role and auth architecture

---

### 2) Superadmin Seeding

Required behavior:
- Create 3 superadmin accounts via a secure seeding path
- Passwords must be hashed and salted
- Store minimal, required identity fields

---

### 3) Access Control

Required behavior:
- Add or extend middleware to restrict superadmin templates and endpoints
- Deny access with a consistent error page or redirect

---

### 4) Superadmin Management UI

Required behavior:
- List superadmin accounts
- Enable or disable an account
- Reset password flow (admin-initiated)
- View audit log entries related to admin actions

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Three superadmin accounts exist and can log in
2) Superadmin pages are not accessible to non-admins
3) Enable/disable and reset password workflows work
4) Admin actions are logged

---

Expected Outcome:
- Secure superadmin accounts and management tools
- Protected admin interfaces aligned with current auth patterns

Priority:
HIGH - Required for governance
