Task: Superadmin Deployment Control (Docker Provisioning)

Objective:
Enable superadmins to view and manage client Docker deployments used to provision HMS instances.

This task must:
- Provide templates to view deployments and their status
- Allow start, stop, restart, and log viewing actions
- Reuse existing provisioning and runner logic
- Enforce superadmin-only access

DO NOT add new dependencies without approval.
DO NOT expose deployment controls to non-admins.

---

### 1) Discovery First: Find Provisioning and Runner Patterns

Before implementation, identify:
- Existing provisioning engine and runner usage
- Where deployment metadata is stored
- Current log collection approach

Implementation rule:
- Extend existing provisioner module and templates

---

### 2) Deployment List View

Required behavior:
- Show deployment name, client, status, and last updated time
- Provide quick actions for start/stop/restart

---

### 3) Deployment Detail View

Required behavior:
- Show environment details and config summary
- Provide log viewer (tail or recent logs)
- Show health or status checks if available

---

### 4) Actions and Logging

Required behavior:
- Implement start/stop/restart using existing runner
- Log all actions in audit history
- Surface errors to the UI

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Superadmin can see all deployments
2) Actions update status correctly
3) Logs are visible without errors
4) Access is restricted to superadmins only

---

Expected Outcome:
- Superadmin deployment dashboard with controls and logs
- No changes to non-admin flows

Priority:
HIGH - Required for provisioning operations
