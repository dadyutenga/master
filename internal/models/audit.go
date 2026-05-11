package models

import "time"

type AuditLog struct {
	ID         int64
	AdminID    int64
	AdminEmail string
	Action     string
	TenantID   *string
	Detail     string
	IPAddress  string
	CreatedAt  time.Time
}
