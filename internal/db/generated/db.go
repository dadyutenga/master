package generated

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	TenantStatusPendingVerification = "pending_verification"
	TenantStatusPendingApproval     = "pending_approval"
	TenantStatusProvisioning        = "provisioning"
	TenantStatusActive              = "active"
	TenantStatusSuspended           = "suspended"
	TenantStatusFailed              = "failed"
)

type CreateUserParams struct {
	Name     string
	Email    string
	Company  string
	Phone    *string
	Password string
}

type CreateTenantParams struct {
	UserID      int64
	CompanyName string
	Slug        string
	Domain      string
	DbName      string
	DbUser      string
	DbPassword  string
}

type UpdateTenantStatusParams struct {
	ID     uuid.UUID
	Status string
}

type SetTenantActiveParams struct {
	ID           uuid.UUID
	AppKey       *string
	ProvisionLog *string
}

type SetTenantFailedParams struct {
	ID           uuid.UUID
	ProvisionLog *string
}

type CreateVerifyTokenParams struct {
	UserID int64
	Token  string
}

type ListTenantsParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	var u User
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO users (name, email, company, phone, password)
		 VALUES (?, ?, ?, ?, ?)
		 RETURNING id, name, email, company, phone, password, role, verified, created_at, updated_at`,
		arg.Name, arg.Email, arg.Company, arg.Phone, arg.Password,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Company, &u.Phone, &u.Password, &u.Role, &u.Verified, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := q.db.QueryRowContext(ctx,
		`SELECT id, name, email, company, phone, password, role, verified, created_at, updated_at FROM users WHERE email = ?`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Company, &u.Phone, &u.Password, &u.Role, &u.Verified, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) GetUserByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := q.db.QueryRowContext(ctx,
		`SELECT id, name, email, company, phone, password, role, verified, created_at, updated_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Company, &u.Phone, &u.Password, &u.Role, &u.Verified, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) VerifyUser(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET verified = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

func (q *Queries) CreateVerifyToken(ctx context.Context, arg CreateVerifyTokenParams) (VerifyToken, error) {
	var vt VerifyToken
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO verify_tokens (user_id, token, expires_at)
		 VALUES (?, ?, datetime('now', '+24 hours'))
		 RETURNING id, user_id, token, expires_at, used`,
		arg.UserID, arg.Token,
	).Scan(&vt.ID, &vt.UserID, &vt.Token, &vt.ExpiresAt, &vt.Used)
	return vt, err
}

func (q *Queries) GetVerifyToken(ctx context.Context, token string) (GetVerifyTokenRow, error) {
	var row GetVerifyTokenRow
	err := q.db.QueryRowContext(ctx,
		`SELECT vt.id, vt.user_id, vt.token, vt.expires_at, vt.used, u.id as uid
		 FROM verify_tokens vt
		 JOIN users u ON vt.user_id = u.id
		 WHERE vt.token = ? AND vt.used = 0 AND vt.expires_at > datetime('now')`,
		token,
	).Scan(&row.ID, &row.UserID, &row.Token, &row.ExpiresAt, &row.Used, &row.Uid)
	return row, err
}

func (q *Queries) UseVerifyToken(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE verify_tokens SET used = 1 WHERE token = ?`,
		token,
	)
	return err
}

func scanTenant(scanner interface {
	Scan(dest ...any) error
}) (Tenant, error) {
	var t Tenant
	var appKey, provisionLog sql.NullString
	var approvedAt, provisionedAt sql.NullTime
	err := scanner.Scan(
		&t.ID, &t.UserID, &t.CompanyName, &t.Slug, &t.Domain,
		&t.DbName, &t.DbUser, &t.DbPassword, &appKey, &t.Status,
		&provisionLog, &approvedAt, &provisionedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return t, err
	}
	if appKey.Valid {
		t.AppKey = &appKey.String
	}
	if provisionLog.Valid {
		t.ProvisionLog = &provisionLog.String
	}
	if approvedAt.Valid {
		t.ApprovedAt = &approvedAt.Time
	}
	if provisionedAt.Valid {
		t.ProvisionedAt = &provisionedAt.Time
	}
	return t, nil
}

func (q *Queries) GetTenantByID(ctx context.Context, id uuid.UUID) (Tenant, error) {
	return scanTenant(q.db.QueryRowContext(ctx,
		`SELECT id, user_id, company_name, slug, domain, db_name, db_user, db_password, app_key, status, provision_log, approved_at, provisioned_at, created_at, updated_at FROM tenants WHERE id = ?`,
		id,
	))
}

func (q *Queries) GetTenantBySlug(ctx context.Context, slug string) (Tenant, error) {
	return scanTenant(q.db.QueryRowContext(ctx,
		`SELECT id, user_id, company_name, slug, domain, db_name, db_user, db_password, app_key, status, provision_log, approved_at, provisioned_at, created_at, updated_at FROM tenants WHERE slug = ?`,
		slug,
	))
}

func (q *Queries) GetTenantByUserID(ctx context.Context, userID int64) (Tenant, error) {
	return scanTenant(q.db.QueryRowContext(ctx,
		`SELECT id, user_id, company_name, slug, domain, db_name, db_user, db_password, app_key, status, provision_log, approved_at, provisioned_at, created_at, updated_at FROM tenants WHERE user_id = ?`,
		userID,
	))
}

func (q *Queries) ListTenants(ctx context.Context, arg ListTenantsParams) ([]ListTenantsRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT t.id, t.user_id, t.company_name, t.slug, t.domain, t.db_name, t.db_user, t.db_password, t.app_key, t.status, t.provision_log, t.approved_at, t.provisioned_at, t.created_at, t.updated_at, u.name as user_name, u.email as user_email
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
		 LIMIT ? OFFSET ?`,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ListTenantsRow
	for rows.Next() {
		var r ListTenantsRow
		var appKey, provisionLog sql.NullString
		var approvedAt, provisionedAt sql.NullTime
		err := rows.Scan(
			&r.ID, &r.UserID, &r.CompanyName, &r.Slug, &r.Domain,
			&r.DbName, &r.DbUser, &r.DbPassword, &appKey, &r.Status,
			&provisionLog, &approvedAt, &provisionedAt, &r.CreatedAt, &r.UpdatedAt,
			&r.UserName, &r.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		if appKey.Valid {
			r.AppKey = &appKey.String
		}
		if provisionLog.Valid {
			r.ProvisionLog = &provisionLog.String
		}
		if approvedAt.Valid {
			r.ApprovedAt = &approvedAt.Time
		}
		if provisionedAt.Valid {
			r.ProvisionedAt = &provisionedAt.Time
		}
		result = append(result, r)
	}
	return result, nil
}

func (q *Queries) CreateTenant(ctx context.Context, arg CreateTenantParams) (Tenant, error) {
	id := uuid.New()
	now := time.Now()
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO tenants (id, user_id, company_name, slug, domain, db_name, db_user, db_password, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending_verification', ?, ?)`,
		id, arg.UserID, arg.CompanyName, arg.Slug, arg.Domain,
		arg.DbName, arg.DbUser, arg.DbPassword, now, now,
	)
	if err != nil {
		return Tenant{}, err
	}
	return q.GetTenantByID(ctx, id)
}

func (q *Queries) UpdateTenantStatus(ctx context.Context, arg UpdateTenantStatusParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		arg.Status, arg.ID,
	)
	return err
}

func (q *Queries) ApproveTenant(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants SET status = 'provisioning', approved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

func (q *Queries) SetTenantActive(ctx context.Context, arg SetTenantActiveParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants SET status = 'active', app_key = ?, provision_log = ?, provisioned_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		arg.AppKey, arg.ProvisionLog, arg.ID,
	)
	return err
}

func (q *Queries) SetTenantFailed(ctx context.Context, arg SetTenantFailedParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants SET status = 'failed', provision_log = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		arg.ProvisionLog, arg.ID,
	)
	return err
}

func (q *Queries) SetTenantProvisioning(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants SET status = 'provisioning', provision_log = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

func (q *Queries) CreateSuperadminIfNotExists(ctx context.Context, arg CreateUserParams) error {
	var count int
	err := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = 'superadmin'`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("Superadmin already exists, skipping.")
		return nil
	}
	_, err = q.db.ExecContext(ctx,
		`INSERT INTO users (name, email, company, phone, password, role, verified)
		 VALUES (?, ?, ?, ?, ?, 'superadmin', 1)`,
		arg.Name, arg.Email, arg.Company, arg.Phone, arg.Password,
	)
	return err
}