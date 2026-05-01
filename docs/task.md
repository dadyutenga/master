## Mega Tasks (5)

### 1) Client Registration + Document Upload Flow
**Goal:** Build a complete client onboarding flow that captures legal, hotel, and personal details, and accepts required documents as images.

**Detailed prompt:**
- Create a multi-step client registration form with validation and draft-save support.
- Required fields: TIN number, BRELA number, hotel name, hotel location (country, city, address), hotel category, room count, and contact info (name, phone, email).
- Upload inputs for BRELA certificate and TRA certificate (images only). Add file size and type validation, preview thumbnails, and an upload progress indicator.
- Persist form data and uploads to the database and object storage (or local file storage), linked to a new Client entity and a VerificationRequest entity.
- Add backend endpoints for create/update registration, upload documents, and submit for verification.
- Ensure server-side validation and sanitize all inputs. Add clear error messages in the UI.

**Acceptance criteria:**
- Client can complete registration and submit documents in one flow.
- Data is stored with a status of "pending_verification".
- Document uploads are visible to the superadmin in the review screen.

---

### 2) Superadmin Accounts + Admin Management Templates
**Goal:** Create three secure superadmin accounts and build templates for superadmin management.

**Detailed prompt:**
- Seed three superadmin accounts with strong passwords and role-based permissions.
- Store credentials securely (hashed, salted). Never hardcode plaintext passwords in source code.
- Add a superadmin management UI: list admins, enable/disable, reset password, and view audit log entries.
- Add an access control middleware to restrict superadmin templates and endpoints.

**Acceptance criteria:**
- Three superadmins exist and can log in securely.
- Superadmin-only pages are protected and non-admins are denied access.

---

### 3) Client Verification Portal + KYC Pages
**Goal:** Provide client-facing pages for verification status and KYC steps after registration.

**Detailed prompt:**
- Add client dashboard pages showing verification status: pending, approved, rejected, and required actions.
- Add KYC pages to collect additional identity and business details after email/password registration.
- Support document re-uploads if a verification request is rejected.
- Send status notifications (UI banners and optional email hooks).

**Acceptance criteria:**
- Client can see current verification status and next steps.
- KYC flow can be completed and resubmitted if needed.

---

### 4) Superadmin Client Verification Workflow
**Goal:** Enable superadmins to review, approve, or reject client registrations with full visibility into documents.

**Detailed prompt:**
- Build a review queue for pending verification requests.
- Provide a detailed view showing all submitted data and document previews.
- Add actions: approve, reject with reason, request more info.
- Record audit logs for all review actions.

**Acceptance criteria:**
- Superadmin can approve or reject requests with a reason.
- Client status updates immediately after action.

---

### 5) Superadmin Deployment Control (Docker Provisioning)
**Goal:** Let superadmins manage Docker-based client instance deployments from the admin panel.

**Detailed prompt:**
- Add templates and backend endpoints to view running deployments and provision new client instances.
- Provide actions to start, stop, restart, and view logs for a client deployment.
- Show deployment status, environment details, and last updated time.
- Ensure only superadmins can access deployment controls.

**Acceptance criteria:**
- Superadmin can view and control deployments from the UI.
- Deployment actions are logged and visible in the admin panel.
