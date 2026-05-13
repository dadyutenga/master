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

	BillingStatusUnpaid    = "unpaid"
	BillingStatusPaid      = "paid"
	BillingStatusOverdue   = "overdue"
	BillingStatusSuspended = "suspended"

	TxnTypeCharge     = "charge"
	TxnTypePayment    = "payment"
	TxnTypeRefund     = "refund"
	TxnTypeAdjustment = "adjustment"

	TxnStatusPending   = "pending"
	TxnStatusCompleted = "completed"
	TxnStatusFailed    = "failed"
	TxnStatusRefunded  = "refunded"

	DeploymentStatusProvisioning = "provisioning"
	DeploymentStatusActive       = "active"
	DeploymentStatusFailed       = "failed"
	DeploymentStatusStopped      = "stopped"
)

// ── Params ──────────────────────────────────────────────────────────────────

type CreateUserParams struct {
	Name        string
	Email       string
	Company     string
	Phone       *string
	Password    string
	TIN         *string
	BrelaNumber *string
}

type UpdateUserParams struct {
	ID          int64
	TIN         *string
	BrelaNumber *string
}

type UpdateUserPasswordParams struct {
	ID       int64
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
	HotelName   *string
	Category    *string
	RoomCount   *int64
	Address     *string
	City        *string
	Country     *string
}

type UpdateTenantDetailsParams struct {
	ID                 uuid.UUID
	RequestedSubdomain *string
	AdminName          *string
	AdminEmail         *string
	AdminPhone         *string
}

type UpdateTenantBillingParams struct {
	ID            uuid.UUID
	BillingStatus string
	LastPaymentAt *time.Time
	NextDueAt     *time.Time
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

type CreateDocumentParams struct {
	UserID       int64
	TenantID     *string
	DocType      string
	Filename     string
	OriginalName string
	MimeType     string
	SizeBytes    int64
}

type CreateContactDetailsParams struct {
	Location    string
	PhoneNumber string
}

type CreateDeploymentParams struct {
	TenantID     uuid.UUID
	Action       string
	Status       string
	Log          *string
	ErrorMessage *string
	CompletedAt  *time.Time
}

type UpdateDeploymentStatusParams struct {
	ID           uuid.UUID
	Status       string
	Log          *string
	ErrorMessage *string
	CompletedAt  *time.Time
}

type UpdateLatestDeploymentStatusParams struct {
	TenantID     uuid.UUID
	Status       string
	Log          *string
	ErrorMessage *string
	CompletedAt  *time.Time
}

// ── Users ───────────────────────────────────────────────────────────────────

func scanUser(scanner interface{ Scan(dest ...any) error }) (User, error) {
	var u User
	var phone, tin, brela sql.NullString
	err := scanner.Scan(
		&u.ID, &u.Name, &u.Email, &u.Company, &phone, &u.Password,
		&u.Role, &u.Verified, &tin, &brela, &u.CreatedAt, &u.UpdatedAt,
	)
	if phone.Valid {
		u.Phone = &phone.String
	}
	if tin.Valid {
		u.TIN = &tin.String
	}
	if brela.Valid {
		u.BrelaNumber = &brela.String
	}
	return u, err
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	var u User
	var phone, tin, brela sql.NullString
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO users (name, email, company, phone, password, tin, brela_number)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 RETURNING id, name, email, company, phone, password, role, verified, tin, brela_number, created_at, updated_at`,
		arg.Name, arg.Email, arg.Company, arg.Phone, arg.Password, arg.TIN, arg.BrelaNumber,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Company, &phone, &u.Password, &u.Role, &u.Verified, &tin, &brela, &u.CreatedAt, &u.UpdatedAt)
	if phone.Valid {
		u.Phone = &phone.String
	}
	if tin.Valid {
		u.TIN = &tin.String
	}
	if brela.Valid {
		u.BrelaNumber = &brela.String
	}
	return u, err
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(q.db.QueryRowContext(ctx,
		`SELECT id, name, email, company, phone, password, role, verified, tin, brela_number, created_at, updated_at FROM users WHERE email = ?`,
		email,
	))
}

func (q *Queries) GetUserByID(ctx context.Context, id int64) (User, error) {
	return scanUser(q.db.QueryRowContext(ctx,
		`SELECT id, name, email, company, phone, password, role, verified, tin, brela_number, created_at, updated_at FROM users WHERE id = ?`,
		id,
	))
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET tin = ?, brela_number = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		arg.TIN, arg.BrelaNumber, arg.ID,
	)
	return err
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		arg.Password, arg.ID,
	)
	return err
}

func (q *Queries) VerifyUser(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE users SET verified = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

// ── Verify Tokens ───────────────────────────────────────────────────────────

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

// ── Contact Details ─────────────────────────────────────────────────────────

func scanContactDetails(scanner interface{ Scan(dest ...any) error }) (ContactDetails, error) {
	var cd ContactDetails
	err := scanner.Scan(&cd.ID, &cd.Location, &cd.PhoneNumber, &cd.CreatedAt, &cd.UpdatedAt)
	return cd, err
}

func (q *Queries) GetContactDetails(ctx context.Context) (ContactDetails, error) {
	return scanContactDetails(q.db.QueryRowContext(ctx,
		`SELECT id, location, phone_number, created_at, updated_at FROM contact_details ORDER BY id LIMIT 1`,
	))
}

func (q *Queries) UpsertContactDetails(ctx context.Context, arg CreateContactDetailsParams) (ContactDetails, error) {
	var cd ContactDetails
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO contact_details (id, location, phone_number, created_at, updated_at)
		 VALUES (1, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT(id) DO UPDATE SET location = excluded.location, phone_number = excluded.phone_number, updated_at = CURRENT_TIMESTAMP
		 RETURNING id, location, phone_number, created_at, updated_at`,
		arg.Location, arg.PhoneNumber,
	).Scan(&cd.ID, &cd.Location, &cd.PhoneNumber, &cd.CreatedAt, &cd.UpdatedAt)
	return cd, err
}

// ── Tenants ─────────────────────────────────────────────────────────────────

func scanTenant(scanner interface{ Scan(dest ...any) error }) (Tenant, error) {
	var t Tenant
	var appKey, provisionLog sql.NullString
	var approvedAt, provisionedAt sql.NullTime
	var hotelName, category, address, city, country sql.NullString
	var requestedSubdomain, adminName, adminEmail, adminPhone, billingStatus sql.NullString
	var roomCount sql.NullInt64
	var lastPaymentAt, nextDueAt sql.NullTime
	err := scanner.Scan(
		&t.ID, &t.UserID, &t.CompanyName, &t.Slug, &t.Domain,
		&t.DbName, &t.DbUser, &t.DbPassword, &appKey, &t.Status,
		&provisionLog, &approvedAt, &provisionedAt,
		&hotelName, &category, &roomCount, &address, &city, &country,
		&requestedSubdomain, &adminName, &adminEmail, &adminPhone, &billingStatus,
		&lastPaymentAt, &nextDueAt,
		&t.CreatedAt, &t.UpdatedAt,
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
	if hotelName.Valid {
		t.HotelName = &hotelName.String
	}
	if category.Valid {
		t.Category = &category.String
	}
	if roomCount.Valid {
		t.RoomCount = &roomCount.Int64
	}
	if address.Valid {
		t.Address = &address.String
	}
	if city.Valid {
		t.City = &city.String
	}
	if country.Valid {
		t.Country = &country.String
	}
	if requestedSubdomain.Valid {
		t.RequestedSubdomain = &requestedSubdomain.String
	}
	if adminName.Valid {
		t.AdminName = &adminName.String
	}
	if adminEmail.Valid {
		t.AdminEmail = &adminEmail.String
	}
	if adminPhone.Valid {
		t.AdminPhone = &adminPhone.String
	}
	if billingStatus.Valid {
		t.BillingStatus = billingStatus.String
	} else {
		t.BillingStatus = BillingStatusUnpaid
	}
	if lastPaymentAt.Valid {
		t.LastPaymentAt = &lastPaymentAt.Time
	}
	if nextDueAt.Valid {
		t.NextDueAt = &nextDueAt.Time
	}
	return t, nil
}

const tenantCols = `id, user_id, company_name, slug, domain, db_name, db_user, db_password, app_key, status, provision_log, approved_at, provisioned_at, hotel_name, category, room_count, address, city, country, requested_subdomain, admin_name, admin_email, admin_phone, billing_status, last_payment_at, next_due_at, created_at, updated_at`

func (q *Queries) GetTenantByID(ctx context.Context, id uuid.UUID) (Tenant, error) {
	return scanTenant(q.db.QueryRowContext(ctx, `SELECT `+tenantCols+` FROM tenants WHERE id = ?`, id))
}

func (q *Queries) GetTenantBySlug(ctx context.Context, slug string) (Tenant, error) {
	return scanTenant(q.db.QueryRowContext(ctx, `SELECT `+tenantCols+` FROM tenants WHERE slug = ?`, slug))
}

func (q *Queries) GetTenantByUserID(ctx context.Context, userID int64) (Tenant, error) {
	return scanTenant(q.db.QueryRowContext(ctx, `SELECT `+tenantCols+` FROM tenants WHERE user_id = ?`, userID))
}

func (q *Queries) GetTenantByRequestedSubdomain(ctx context.Context, subdomain string) (Tenant, error) {
	return scanTenant(q.db.QueryRowContext(ctx, `SELECT `+tenantCols+` FROM tenants WHERE requested_subdomain = ?`, subdomain))
}

func (q *Queries) ListTenants(ctx context.Context, arg ListTenantsParams) ([]ListTenantsRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT t.id, t.user_id, t.company_name, t.slug, t.domain, t.db_name, t.db_user, t.db_password, t.app_key, t.status, t.provision_log, t.approved_at, t.provisioned_at, t.billing_status, t.created_at, t.updated_at, u.name as user_name, u.email as user_email
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
			&provisionLog, &approvedAt, &provisionedAt, &r.BillingStatus, &r.CreatedAt, &r.UpdatedAt,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (q *Queries) CreateTenant(ctx context.Context, arg CreateTenantParams) (Tenant, error) {
	id := uuid.New()
	now := time.Now()
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO tenants (id, user_id, company_name, slug, domain, db_name, db_user, db_password, status, hotel_name, category, room_count, address, city, country, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending_verification', ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, arg.UserID, arg.CompanyName, arg.Slug, arg.Domain,
		arg.DbName, arg.DbUser, arg.DbPassword, arg.HotelName, arg.Category, arg.RoomCount, arg.Address, arg.City, arg.Country, now, now,
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

func (q *Queries) UpdateTenantDetails(ctx context.Context, arg UpdateTenantDetailsParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants
		 SET requested_subdomain = ?, admin_name = ?, admin_email = ?, admin_phone = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		arg.RequestedSubdomain, arg.AdminName, arg.AdminEmail, arg.AdminPhone, arg.ID,
	)
	return err
}

func (q *Queries) UpdateTenantBilling(ctx context.Context, arg UpdateTenantBillingParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants
		 SET billing_status = ?, last_payment_at = ?, next_due_at = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		arg.BillingStatus, arg.LastPaymentAt, arg.NextDueAt, arg.ID,
	)
	return err
}

func (q *Queries) ListOverdueTenants(ctx context.Context) ([]Tenant, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+tenantCols+` FROM tenants
		 WHERE billing_status = 'paid' AND next_due_at IS NOT NULL AND next_due_at < datetime('now')`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func (q *Queries) MarkTenantOverdue(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants SET billing_status = 'overdue', updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
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

// ── Deployments ─────────────────────────────────────────────────────────────

func scanDeployment(scanner interface{ Scan(dest ...any) error }) (Deployment, error) {
	var d Deployment
	var logText, errText sql.NullString
	var completedAt sql.NullTime
	err := scanner.Scan(
		&d.ID, &d.TenantID, &d.Action, &d.Status, &logText, &errText,
		&d.CreatedAt, &d.UpdatedAt, &completedAt,
	)
	if logText.Valid {
		d.Log = &logText.String
	}
	if errText.Valid {
		d.ErrorMessage = &errText.String
	}
	if completedAt.Valid {
		d.CompletedAt = &completedAt.Time
	}
	return d, err
}

func (q *Queries) CreateDeployment(ctx context.Context, arg CreateDeploymentParams) (Deployment, error) {
	id := uuid.New()
	return scanDeployment(q.db.QueryRowContext(ctx,
		`INSERT INTO deployments (id, tenant_id, action, status, log, error_message, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 RETURNING id, tenant_id, action, status, log, error_message, created_at, updated_at, completed_at`,
		id, arg.TenantID, arg.Action, arg.Status, arg.Log, arg.ErrorMessage, arg.CompletedAt,
	))
}

func (q *Queries) GetDeploymentByID(ctx context.Context, id uuid.UUID) (Deployment, error) {
	return scanDeployment(q.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, action, status, log, error_message, created_at, updated_at, completed_at FROM deployments WHERE id = ?`,
		id,
	))
}

func (q *Queries) ListDeploymentsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]Deployment, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, tenant_id, action, status, log, error_message, created_at, updated_at, completed_at
		 FROM deployments WHERE tenant_id = ? ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Deployment
	for rows.Next() {
		d, err := scanDeployment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (q *Queries) UpdateDeploymentStatus(ctx context.Context, arg UpdateDeploymentStatusParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE deployments SET status = ?, log = ?, error_message = ?, completed_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		arg.Status, arg.Log, arg.ErrorMessage, arg.CompletedAt, arg.ID,
	)
	return err
}

func (q *Queries) UpdateLatestDeploymentStatus(ctx context.Context, arg UpdateLatestDeploymentStatusParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE deployments
		 SET status = ?, log = ?, error_message = ?, completed_at = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = (SELECT id FROM deployments WHERE tenant_id = ? ORDER BY created_at DESC LIMIT 1)`,
		arg.Status, arg.Log, arg.ErrorMessage, arg.CompletedAt, arg.TenantID,
	)
	return err
}

// ── Documents ──────────────────────────────────────────────────────────────

func (q *Queries) CreateDocument(ctx context.Context, arg CreateDocumentParams) (Document, error) {
	var d Document
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO documents (user_id, tenant_id, doc_type, filename, original_name, mime_type, size_bytes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 RETURNING id, user_id, tenant_id, doc_type, filename, original_name, mime_type, size_bytes, created_at`,
		arg.UserID, arg.TenantID, arg.DocType, arg.Filename, arg.OriginalName, arg.MimeType, arg.SizeBytes,
	).Scan(&d.ID, &d.UserID, &d.TenantID, &d.DocType, &d.Filename, &d.OriginalName, &d.MimeType, &d.SizeBytes, &d.CreatedAt)
	return d, err
}

func (q *Queries) GetDocumentsByUserID(ctx context.Context, userID int64) ([]Document, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, tenant_id, doc_type, filename, original_name, mime_type, size_bytes, created_at FROM documents WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.UserID, &d.TenantID, &d.DocType, &d.Filename, &d.OriginalName, &d.MimeType, &d.SizeBytes, &d.CreatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return docs, nil
}

func (q *Queries) DeleteDocument(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
	return err
}

// ── Admin ────────────────────────────────────────────────────────────────────

func (q *Queries) CreateSuperadminIfNotExists(ctx context.Context, arg CreateUserParams) error {
	var count int
	err := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("Admin already exists, skipping.")
		return nil
	}
	_, err = q.db.ExecContext(ctx,
		`INSERT INTO users (name, email, company, phone, password, role, verified)
		 VALUES (?, ?, ?, ?, ?, 'admin', 1)`,
		arg.Name, arg.Email, arg.Company, arg.Phone, arg.Password,
	)
	return err
}

func (q *Queries) CreateBillingTransaction(ctx context.Context, tenantID string, amount float64, txnType, description string, adminID *int64) (BillingTransaction, error) {
	result, err := q.db.ExecContext(ctx,
		`INSERT INTO billing_transactions (tenant_id, amount, transaction_type, description, admin_id)
		 VALUES (?, ?, ?, ?, ?)`,
		tenantID, amount, txnType, description, adminID,
	)
	if err != nil {
		return BillingTransaction{}, err
	}
	id, _ := result.LastInsertId()
	return BillingTransaction{
		ID: id, TenantID: tenantID, Amount: amount, Currency: "TZS",
		TransactionType: txnType, Description: description, Status: TxnStatusCompleted,
		AdminID: adminID, CreatedAt: time.Now(),
	}, nil
}

func (q *Queries) GetBillingTransactionsByTenantID(ctx context.Context, tenantID string) ([]BillingTransaction, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, tenant_id, amount, currency, description, transaction_type, status, admin_id, created_at
		 FROM billing_transactions WHERE tenant_id = ? ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txns []BillingTransaction
	for rows.Next() {
		var t BillingTransaction
		var adminID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Amount, &t.Currency, &t.Description, &t.TransactionType, &t.Status, &adminID, &t.CreatedAt); err != nil {
			return nil, err
		}
		if adminID.Valid {
			t.AdminID = &adminID.Int64
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

func (q *Queries) GetBillingTransactionByID(ctx context.Context, id int64) (BillingTransaction, error) {
	var t BillingTransaction
	var adminID sql.NullInt64
	err := q.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, amount, currency, description, transaction_type, status, admin_id, created_at
		 FROM billing_transactions WHERE id = ?`, id,
	).Scan(&t.ID, &t.TenantID, &t.Amount, &t.Currency, &t.Description, &t.TransactionType, &t.Status, &adminID, &t.CreatedAt)
	if adminID.Valid {
		t.AdminID = &adminID.Int64
	}
	return t, err
}

func (q *Queries) MarkTenantPaid(ctx context.Context, tenantID string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE tenants SET billing_status = 'paid', last_payment_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		tenantID)
	return err
}

// ── Instances ──────────────────────────────────────────────────────────────

const instanceCols = `id, tenant_id, hotel_name, slug, domain, db_name, db_user, db_password, app_key, status, admin_disabled, billing_status, price, package_name, last_payment_at, next_due_at, provision_log, approved_at, provisioned_at, archived_at, deleted_at, created_at, updated_at`

func scanInstance(scanner interface{ Scan(dest ...any) error }) (Instance, error) {
	var inst Instance
	var appKey, provisionLog sql.NullString
	var lastPaymentAt, nextDueAt, approvedAt, provisionedAt, archivedAt, deletedAt sql.NullTime
	var price sql.NullFloat64
	err := scanner.Scan(
		&inst.ID, &inst.TenantID, &inst.HotelName, &inst.Slug, &inst.Domain,
		&inst.DbName, &inst.DbUser, &inst.DbPassword, &appKey, &inst.Status,
		&inst.AdminDisabled, &inst.BillingStatus, &price, &inst.PackageName,
		&lastPaymentAt, &nextDueAt, &provisionLog,
		&approvedAt, &provisionedAt, &archivedAt, &deletedAt,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return inst, err
	}
	if appKey.Valid {
		inst.AppKey = &appKey.String
	}
	if provisionLog.Valid {
		inst.ProvisionLog = &provisionLog.String
	}
	if lastPaymentAt.Valid {
		inst.LastPaymentAt = &lastPaymentAt.Time
	}
	if nextDueAt.Valid {
		inst.NextDueAt = &nextDueAt.Time
	}
	if approvedAt.Valid {
		inst.ApprovedAt = &approvedAt.Time
	}
	if provisionedAt.Valid {
		inst.ProvisionedAt = &provisionedAt.Time
	}
	if archivedAt.Valid {
		inst.ArchivedAt = &archivedAt.Time
	}
	if deletedAt.Valid {
		inst.DeletedAt = &deletedAt.Time
	}
	if price.Valid {
		inst.Price = price.Float64
	}
	return inst, nil
}

func (q *Queries) GetInstanceByID(ctx context.Context, id uuid.UUID) (Instance, error) {
	return scanInstance(q.db.QueryRowContext(ctx, `SELECT `+instanceCols+` FROM instances WHERE id = ?`, id))
}

func (q *Queries) ListInstancesByTenantID(ctx context.Context, tenantID uuid.UUID) ([]Instance, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+instanceCols+` FROM instances WHERE tenant_id = ? AND deleted_at IS NULL ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var instances []Instance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (q *Queries) ListAllInstances(ctx context.Context) ([]Instance, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+instanceCols+` FROM instances WHERE deleted_at IS NULL ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var instances []Instance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (q *Queries) ListArchivedInstances(ctx context.Context) ([]Instance, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+instanceCols+` FROM instances WHERE archived_at IS NOT NULL AND deleted_at IS NULL ORDER BY archived_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var instances []Instance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (q *Queries) CreateInstance(ctx context.Context, tenantID uuid.UUID, hotelName, slug, dbName, dbUser, dbPassword, packageName string, price float64) (Instance, error) {
	id := uuid.New()
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO instances (id, tenant_id, hotel_name, slug, domain, db_name, db_user, db_password, status, billing_status, price, package_name)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending_payment', 'unpaid', ?, ?)`,
		id, tenantID, hotelName, slug, slug+".hms.local", dbName, dbUser, dbPassword, price, packageName,
	)
	if err != nil {
		return Instance{}, err
	}
	return q.GetInstanceByID(ctx, id)
}

func (q *Queries) UpdateInstanceStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE instances SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	return err
}

func (q *Queries) SetInstanceAdminDisabled(ctx context.Context, id uuid.UUID, disabled bool) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE instances SET admin_disabled = ?, status = CASE WHEN ? = 1 THEN 'disabled' ELSE 'active' END, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		disabled, disabled, id,
	)
	return err
}

func (q *Queries) ArchiveInstance(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE instances SET status = 'archived', archived_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

func (q *Queries) DeleteInstance(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE instances SET status = 'deleted', deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

func (q *Queries) UpdateInstanceBilling(ctx context.Context, id uuid.UUID, billingStatus string, lastPayment, nextDue *time.Time) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE instances SET billing_status = ?, last_payment_at = ?, next_due_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		billingStatus, lastPayment, nextDue, id,
	)
	return err
}

func (q *Queries) UpdateInstancePrice(ctx context.Context, id uuid.UUID, price float64, packageName string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE instances SET price = ?, package_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		price, packageName, id,
	)
	return err
}

func (q *Queries) MarkInstancePaid(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE instances SET billing_status = 'paid', last_payment_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

func (q *Queries) GetInstanceBySlug(ctx context.Context, slug string) (Instance, error) {
	return scanInstance(q.db.QueryRowContext(ctx, `SELECT `+instanceCols+` FROM instances WHERE slug = ?`, slug))
}

func (q *Queries) CountInstancesByTenantID(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	err := q.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM instances WHERE tenant_id = ? AND deleted_at IS NULL`,
		tenantID,
	).Scan(&count)
	return count, err
}

func (q *Queries) CountInstancesByStatus(ctx context.Context) (active, paused, disabled, failed int, err error) {
	err = q.db.QueryRowContext(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN status='active' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='paused' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='disabled' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='failed' THEN 1 ELSE 0 END), 0)
		FROM instances WHERE deleted_at IS NULL`,
	).Scan(&active, &paused, &disabled, &failed)
	return
}

func (q *Queries) ListOverdueInstances(ctx context.Context) ([]Instance, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT `+instanceCols+` FROM instances
		 WHERE billing_status = 'paid' AND next_due_at IS NOT NULL AND next_due_at < datetime('now') AND deleted_at IS NULL`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var instances []Instance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// ── Instance Deployments ──────────────────────────────────────────────────

func scanInstanceDeployment(scanner interface{ Scan(dest ...any) error }) (InstanceDeployment, error) {
	var d InstanceDeployment
	var logText, errText sql.NullString
	var completedAt sql.NullTime
	err := scanner.Scan(
		&d.ID, &d.InstanceID, &d.Action, &d.Status, &logText, &errText,
		&d.CreatedAt, &d.UpdatedAt, &completedAt,
	)
	if logText.Valid {
		d.Log = &logText.String
	}
	if errText.Valid {
		d.ErrorMessage = &errText.String
	}
	if completedAt.Valid {
		d.CompletedAt = &completedAt.Time
	}
	return d, err
}

func (q *Queries) CreateInstanceDeployment(ctx context.Context, instanceID uuid.UUID, action, status string) (InstanceDeployment, error) {
	id := uuid.New()
	return scanInstanceDeployment(q.db.QueryRowContext(ctx,
		`INSERT INTO instance_deployments (id, instance_id, action, status)
		 VALUES (?, ?, ?, ?)
		 RETURNING id, instance_id, action, status, log, error_message, created_at, updated_at, completed_at`,
		id, instanceID, action, status,
	))
}

func (q *Queries) ListInstanceDeployments(ctx context.Context, instanceID uuid.UUID) ([]InstanceDeployment, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, instance_id, action, status, log, error_message, created_at, updated_at, completed_at
		 FROM instance_deployments WHERE instance_id = ? ORDER BY created_at DESC`,
		instanceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []InstanceDeployment
	for rows.Next() {
		d, err := scanInstanceDeployment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

// ── Billing Packages ──────────────────────────────────────────────────────

func (q *Queries) ListBillingPackages(ctx context.Context) ([]BillingPackage, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, description, price, currency, billing_cycle, is_active, created_at, updated_at
		 FROM billing_packages WHERE is_active = 1 ORDER BY price ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var packages []BillingPackage
	for rows.Next() {
		var p BillingPackage
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Currency, &p.BillingCycle, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, rows.Err()
}

func (q *Queries) GetAllBillingPackages(ctx context.Context) ([]BillingPackage, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, description, price, currency, billing_cycle, is_active, created_at, updated_at
		 FROM billing_packages ORDER BY price ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var packages []BillingPackage
	for rows.Next() {
		var p BillingPackage
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Currency, &p.BillingCycle, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, rows.Err()
}

func (q *Queries) GetBillingPackageByID(ctx context.Context, id int64) (BillingPackage, error) {
	var p BillingPackage
	err := q.db.QueryRowContext(ctx,
		`SELECT id, name, description, price, currency, billing_cycle, is_active, created_at, updated_at
		 FROM billing_packages WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Currency, &p.BillingCycle, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (q *Queries) CreateBillingPackage(ctx context.Context, name, description string, price float64, currency, billingCycle string) (BillingPackage, error) {
	result, err := q.db.ExecContext(ctx,
		`INSERT INTO billing_packages (name, description, price, currency, billing_cycle)
		 VALUES (?, ?, ?, ?, ?)`,
		name, description, price, currency, billingCycle,
	)
	if err != nil {
		return BillingPackage{}, err
	}
	id, _ := result.LastInsertId()
	return BillingPackage{
		ID: id, Name: name, Description: description, Price: price,
		Currency: currency, BillingCycle: billingCycle, IsActive: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}, nil
}

func (q *Queries) UpdateBillingPackage(ctx context.Context, id int64, name, description string, price float64, billingCycle string, isActive bool) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE billing_packages SET name = ?, description = ?, price = ?, billing_cycle = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, description, price, billingCycle, isActive, id,
	)
	return err
}

func (q *Queries) DeleteBillingPackage(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE billing_packages SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}
