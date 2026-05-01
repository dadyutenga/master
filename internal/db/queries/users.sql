-- name: CreateUser :one
INSERT INTO users (name, email, company, phone, password)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: VerifyUser :exec
UPDATE users SET verified = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: CreateVerifyToken :one
INSERT INTO verify_tokens (user_id, token, expires_at)
VALUES (?, ?, datetime('now', '+24 hours'))
RETURNING *;

-- name: GetVerifyToken :one
SELECT vt.*, u.id as uid FROM verify_tokens vt
JOIN users u ON vt.user_id = u.id
WHERE vt.token = ? AND vt.used = 0 AND vt.expires_at > datetime('now');

-- name: UseVerifyToken :exec
UPDATE verify_tokens SET used = 1 WHERE token = ?;