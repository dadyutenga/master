package models

import "database/sql"

type DashboardStats struct {
	TotalTenants       int
	ActiveTenants      int
	PendingTenants     int
	SuspendedTenants   int
	RunningDeployments int
	FailedDeployments  int
	RecentActions      []AuditLog
}

func LoadDashboardStats(db *sql.DB, audit *AuditStore) (DashboardStats, error) {
	var s DashboardStats

	err := db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status='active' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status IN ('pending_verification','pending_approval') THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='suspended' THEN 1 ELSE 0 END), 0)
		FROM tenants
	`).Scan(&s.TotalTenants, &s.ActiveTenants, &s.PendingTenants, &s.SuspendedTenants)
	if err != nil && err != sql.ErrNoRows {
		return s, err
	}

	_ = db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN status='active' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status='failed' THEN 1 ELSE 0 END), 0)
		FROM deployments
	`).Scan(&s.RunningDeployments, &s.FailedDeployments)

	if audit != nil {
		page, err := audit.List(1, 10, "", "")
		if err != nil {
			return s, err
		}
		s.RecentActions = page.Logs
	}

	return s, nil
}
