Task: Client Registration and Document Upload Flow

Objective:
Create a complete client onboarding flow that captures legal, hotel, and personal details and collects required documents as images. The flow must be secure, validated, and ready for superadmin verification.

This task must:
- Provide a multi-step registration flow for clients
- Collect TIN and BRELA numbers and required supporting documents
- Capture hotel details, location, and primary contact data
- Store data and documents securely with a pending verification status
- Reuse existing view, service, and storage patterns

DO NOT hardcode credentials.
DO NOT bypass server-side validation.
DO NOT change unrelated module flows.

---

### 1) Discovery First: Find Existing Patterns

Before implementation, identify:
- Existing registration flows (client or admin)
- Current file upload handling and storage helpers
- Validation patterns and error handling style
- Any existing verification status fields in the database

Implementation rule:
- Extend current conventions and naming
- Avoid duplicate upload services

---

### 2) Registration Form Structure

Required behavior:
- Multi-step or sectioned form with clear progression
- Draft save support (optional but preferred if patterns exist)

Required fields:
- Legal: TIN number, BRELA number
- Hotel: hotel name, category, room count, address, city, country
- Contact: full name, phone, email

Document uploads (images only):
- BRELA certificate
- TRA certificate

---

### 3) Backend Endpoints and Storage

Required behavior:
- Create or update registration endpoint
- Upload endpoints for documents
- Save data with status "pending_verification"
- Link uploads to the client and verification request

Implementation rule:
- Reuse existing database and storage helpers
- Apply server-side validation for all inputs

---

### 4) UI Validation and UX

Required behavior:
- Client-side validation for required fields and file types
- Show inline errors and upload progress
- Provide thumbnail previews for documents

---

### 5) Security and Access Control

Required behavior:
- Ensure only unauthenticated clients can start registration
- Prevent cross-tenant access to uploads
- Sanitize all inputs

---

### 6) Testing and Manual Acceptance Criteria

Validate end-to-end:
1) Client can submit all required data and documents
2) Documents are stored and linked to the client
3) Status is set to pending verification
4) Invalid files are rejected
5) Server-side validation rejects invalid data

---

Expected Outcome:
- Fully functional client registration and document upload flow
- Data stored consistently and ready for superadmin review

Priority:
HIGH - Required for onboarding
