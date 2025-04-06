package role

type Role string

const (
	DashboardRoleAdmin    Role = "admin"
	DashboardRoleEmployee Role = "employee"
	DashboardRoleManager  Role = "manager"
	DashboardRoleUnknown  Role = "unknown"
)
