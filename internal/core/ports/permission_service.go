package ports

// PermissionService resolves role-based permissions from an in-memory map.
// Loaded at startup from the database, refreshed periodically.
type PermissionService interface {
	HasPermission(role string, permission string) bool
	GetPermissions(role string) []string
}
