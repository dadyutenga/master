Task: Admin Deployment Management per Tenant

Objective:
Allow admins to view deployments for each tenant and perform basic management actions to fix or manage those deployments.

This task must:
- Show a list of deployments for each tenant in the admin area
- Provide basic management actions (view status, retry, stop/start if supported)
- Reuse existing provisioning and deployment patterns

DO NOT expose deployment actions to non-admins.
DO NOT change unrelated provisioning flows.

---

### 1) Discovery First: Find Existing Deployment Patterns

Before implementation, identify:
- Existing provisioning engine and runner usage
- Where deployment status is stored (db tables or logs)
- Admin pages that already show tenant details

Implementation rule:
- Extend existing provisioner and admin handler patterns

---

### 2) Deployment Data Model

Required behavior:
- Each tenant has a list of deployments with status, timestamps, and errors
- Link deployments to tenant IDs
- Store management actions as audit events if pattern exists

---

### 3) Admin UI and Actions

Required behavior:
- Admin can view a tenant deployment list
- Admin can open a deployment detail view
- Admin can trigger management actions (retry, stop, start) if supported by provisioner

Implementation notes:
- Actions must confirm intent before running
+- Show clear status and error messages

---

### 4) Security and Validation

Required behavior:
- Admin-only access enforced for all endpoints
- Validate action inputs and tenant scope

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Admin can list deployments per tenant
2) Admin can view deployment status and errors
3) Admin management actions work and are logged
4) Non-admins cannot access deployment views

---

Expected Outcome:
- Admin can view and manage tenant deployments safely

Priority:
HIGH
