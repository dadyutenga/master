package models

import (
	"database/sql"
	"log"
	"time"
)

type AuditLog struct {
	ID         int64
	AdminID    int64
	AdminEmail string
	Action     string
	TenantID   *string
	TenantName *string
	Detail     string
	IPAddress  string
	CreatedAt  time.Time
}

type AuditPage struct {
	Logs       []AuditLog
	Page       int
	TotalPages int
	Total      int
}

type AuditStore struct {
	db *sql.DB
}

func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

// Log inserts an audit record. Errors are swallowed — never block
// the real action because of an audit failure.
func (s *AuditStore) Log(adminID int64, action string, tenantID *string, detail, ip string) {
	_, err := s.db.Exec(
		`INSERT INTO audit_logs (admin_id, action, tenant_id, detail, ip_address)
		 VALUES (?, ?, ?, ?, ?)`,
		adminID, action, tenantID, detail, ip,
	)
	if err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func (s *AuditStore) List(page, limit int, action, search string) (AuditPage, error) {
	offset := (page - 1) * limit

	base := `FROM audit_logs a
			 LEFT JOIN users   u ON u.id = a.admin_id
			 LEFT JOIN tenants t ON t.id = a.tenant_id
			 WHERE 1=1`
	args := []interface{}{}

	if action != "" {
		base += " AND a.action = ?"
		args = append(args, action)
	}
	if search != "" {
		base += " AND (u.email LIKE ? OR t.company_name LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	var total int
	s.db.QueryRow("SELECT COUNT(*) "+base, args...).Scan(&total)

	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := s.db.Query(
		`SELECT a.id, a.admin_id, COALESCE(u.email,'deleted'),
				a.action, a.tenant_id, t.company_name,
				COALESCE(a.detail,''), COALESCE(a.ip_address,''), a.created_at
		 `+base+` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`,
		queryArgs...,
	)
	if err != nil {
		return AuditPage{}, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		var createdAt string
		rows.Scan(
			&l.ID, &l.AdminID, &l.AdminEmail, &l.Action,
			&l.TenantID, &l.TenantName, &l.Detail, &l.IPAddress, &createdAt,
		)
		l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		logs = append(logs, l)
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	return AuditPage{Logs: logs, Page: page, TotalPages: totalPages, Total: total}, nil
}
