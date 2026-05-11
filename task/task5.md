Task: Billing Module and Access Enforcement

Objective:
Add a billing module that disables tenant access when payment is overdue. If unpaid, the tenant container should be blocked and the tenant subdomain becomes inaccessible until payment.

This task must:
- Add billing status tracking for each tenant
- Enforce access restrictions when unpaid
- Provide admin visibility and tenant notification

DO NOT hardcode billing status.
DO NOT break existing tenant provisioning flows.

---

### 1) Discovery First: Find Tenant Access and Provisioning Patterns

Before implementation, identify:
- How tenant subdomains and routing are handled
- Where container status and access controls are enforced
- Existing billing or subscription fields (if any)

Implementation rule:
- Extend current access control and provisioning patterns

---

### 2) Billing Data Model

Required behavior:
- Track billing status per tenant: paid, overdue, suspended
- Store last payment date and next due date
- Provide queries to update billing status

---

### 3) Enforcement Rules

Required behavior:
- If status is overdue or suspended, tenant subdomain access is blocked
- Container should be stopped or not routed until payment is resolved
- On payment confirmation, restore access and container routing

Implementation notes:
- Use existing middleware or routing layers for enforcement
- Avoid hard stops if a grace period is configured

---

### 4) Admin and Tenant Views

Required behavior:
- Admin can view billing status for each tenant
- Admin can update status after payment
- Tenant sees a billing warning or payment required page when suspended

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Tenant with unpaid status cannot access subdomain
2) Container access is blocked until status is paid
3) Admin can update billing status
4) Tenant access is restored after payment

---

Expected Outcome:
- Billing module enforces access control for unpaid tenants

Priority:
HIGH
