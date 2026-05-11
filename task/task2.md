Task: Verification Page Dialog and Contact Visibility

Objective:
Add a dialog box on the verification page and ensure contact details are visible on that page.

This task must:
- Add a dialog box on the verification page for user guidance or confirmation
- Render contact details on the verification page
- Reuse existing UI components and patterns

DO NOT hardcode contact details.
DO NOT bypass existing validation or auth checks.

---

### 1) Discovery First: Find Existing Verification Page

Before implementation, identify:
- The verification page template and handler
- Existing modal/dialog implementation (if any)
- How contact details are currently displayed elsewhere

Implementation rule:
- Reuse existing UI patterns for dialogs and alerts

---

### 2) Dialog Box Requirements

Required behavior:
- Display a dialog box on the verification page
- Dialog content should explain verification status and next action
- Dialog must be accessible (focus trap, close on ESC if pattern exists)

Implementation notes:
- Use existing modal styles and JS if present
- Ensure dialog does not block essential navigation

---

### 3) Contact Details Visibility

Required behavior:
- Show location and phone number on the verification page
- Values must be fetched from the database
- Update in real time after admin changes (on page reload)

---

### 4) Validation and Security

Required behavior:
- Do not expose other tenant data
- Respect current access control for verification pages

---

### 5) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Verification page shows a dialog box with clear guidance
2) Contact details (location, phone number) are visible
3) Details are data-driven from the database
4) Page renders correctly on desktop and mobile

---

Expected Outcome:
- Verification page includes a dialog and visible contact details

Priority:
HIGH
