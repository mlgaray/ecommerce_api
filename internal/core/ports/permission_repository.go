package ports

import "context"

// PermissionRepository loads role-permission mappings from the database.
type PermissionRepository interface {
	GetAllRolePermissions(ctx context.Context) (map[string][]string, error)
}
