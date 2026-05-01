Task: Superadmin Client Verification Workflow

Objective:
Provide a superadmin review workflow to approve or reject client registrations with full document visibility and audit tracking.

This task must:
- Provide a review queue for pending requests
- Show all submitted data and documents
- Allow approve, reject, and request-more-info actions
- Record audit logs for every action

DO NOT change client data outside the review workflow.
DO NOT bypass permissions.

---

### 1) Discovery First: Find Review and Audit Patterns

Before implementation, identify:
- Existing review or approval queues
- Document viewer or file preview components
- Audit logging conventions

Implementation rule:
- Follow existing naming and data patterns

---

### 2) Review Queue

Required behavior:
- List all pending verification requests
- Sort and filter by date, status, or hotel name
- Show key summary fields in the list

---

### 3) Review Detail View

Required behavior:
- Display all submitted data
- Show document previews with download links
- Provide review actions: approve, reject, request more info

---

### 4) Audit and Status Updates

Required behavior:
- Log action, user, and timestamp
- Update client verification status
- Notify client of decision (UI banner, optional email hook)

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Pending requests appear in the queue
2) Superadmin can approve or reject with reason
3) Status updates are reflected for the client
4) Audit logs include review events

---

Expected Outcome:
- Clear, secure verification workflow for superadmins
- Full visibility and auditability of decisions

Priority:
HIGH - Required for KYC compliance
