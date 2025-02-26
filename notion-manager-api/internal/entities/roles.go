package entities

type DashboardRole string

const (
	DashboardRoleAdmin    DashboardRole = "admin"
	DashboardRoleEmployee DashboardRole = "employee"
	DashboardRoleManager  DashboardRole = "manager"
	DashboardRoleUnknown  DashboardRole = "unknown"
)
