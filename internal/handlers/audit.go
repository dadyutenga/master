package handlers

import (
	"database/sql"
	"log"
)

func LogAction(db *sql.DB, adminID int64, action string, tenantID *string, detail, ip string) {
	_, err := db.Exec(
		`INSERT INTO audit_logs (admin_id, action, tenant_id, detail, ip_address)
		 VALUES (?, ?, ?, ?, ?)`,
		adminID, action, tenantID, detail, ip,
	)
	if err != nil {
		log.Printf("audit log error: %v", err)
	}
}
