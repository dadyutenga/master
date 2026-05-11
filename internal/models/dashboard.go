package models

import "time"

type DashboardStats struct {
	TotalTenants       int
	ActiveTenants      int
	PendingTenants     int
	SuspendedTenants   int
	RunningDeployments int
	FailedDeployments  int
	RecentActions      []AuditLog
}

type TenantRow struct {
	ID            string
	CompanyName   string
	Email         string
	Status        string
	BillingStatus string
	Domain        string
	CreatedAt     time.Time
}

type TenantPage struct {
	Tenants    []TenantRow
	Page       int
	TotalPages int
	Total      int
	Status     string
	Search     string
}
