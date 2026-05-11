package models

type DashboardStats struct {
	TotalTenants       int
	ActiveTenants      int
	PendingTenants     int
	SuspendedTenants   int
	RunningDeployments int
	FailedDeployments  int
	RecentActions      []AuditLog
}
