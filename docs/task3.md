Task: Client Verification Portal and KYC Pages

Objective:
Build client-facing verification and KYC pages that show status and guide clients through required steps after registration.

This task must:
- Show verification status and next actions
- Provide KYC data capture after email and password registration
- Allow document re-upload when rejected
- Reuse existing view and validation patterns

DO NOT remove existing registration flows.
DO NOT expose other tenants' data.

---

### 1) Discovery First: Find Existing Client Dashboards

Before implementation, identify:
- Existing client dashboard or profile templates
- Status or verification fields in database models
- Notification patterns (UI or email)

Implementation rule:
- Extend current client view conventions

---

### 2) Verification Status Pages

Required behavior:
- Status display: pending, approved, rejected, more_info_required
- Show timestamps and reason for rejection if present
- Provide a clear call-to-action for next steps

---

### 3) KYC Data Collection

Required behavior:
- Collect additional identity and business data
- Validate and store KYC data linked to the client
- Support a resubmission flow if rejected

---

### 4) Document Re-Upload

Required behavior:
- Allow re-upload of rejected documents
- Keep previous versions in audit history (if pattern exists)

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Client can view verification status at any time
2) KYC pages can be completed and saved
3) Rejected clients can resubmit documents
4) Status updates are visible immediately

---

Expected Outcome:
- Client verification portal with full KYC workflow
- Clear status visibility and next steps

Priority:
HIGH - Required for onboarding completion
