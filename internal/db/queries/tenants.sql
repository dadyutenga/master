-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = ?;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = ?;

-- name: GetTenantByUserID :one
SELECT * FROM tenants WHERE user_id = ?;

-- name: ListTenants :many
SELECT t.*, u.name as user_name, u.email as user_email
FROM tenants t
JOIN users u ON t.user_id = u.id
ORDER BY
    CASE t.status
        WHEN 'pending_approval' THEN 1
        WHEN 'provisioning'     THEN 2
        WHEN 'failed'           THEN 3
        WHEN 'active'           THEN 4
        ELSE 5
    END,
    t.created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateTenant :one
INSERT INTO tenants (id, user_id, company_name, slug, domain, db_name, db_user, db_password, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending_verification')
RETURNING *;

-- name: UpdateTenantStatus :exec
UPDATE tenants SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ApproveTenant :exec
UPDATE tenants SET status = 'provisioning', approved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: SetTenantActive :exec
UPDATE tenants SET status = 'active', app_key = ?, provision_log = ?, provisioned_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: SetTenantFailed :exec
UPDATE tenants SET status = 'failed', provision_log = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: SetTenantProvisioning :exec
UPDATE tenants SET status = 'provisioning', provision_log = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?;