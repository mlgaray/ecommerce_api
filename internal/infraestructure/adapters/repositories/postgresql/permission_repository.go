package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const (
	PermissionRepositoryField = "permission_repository"
)

type PermissionSQLRepository struct {
	db *sql.DB
}

func NewPermissionRepository(dataBaseConnection DataBaseConnection) ports.PermissionRepository {
	return &PermissionSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

// GetAllRolePermissions loads all role → permission mappings from the database.
func (r *PermissionSQLRepository) GetAllRolePermissions(ctx context.Context) (map[string][]string, error) {
	const query = `
		SELECT r.name as role, p.name as permission
		FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		ORDER BY r.name, p.name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     PermissionRepositoryField,
			"function": "get_all_role_permissions",
			"error":    err.Error(),
		}).Error("Failed to load role permissions")
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var role, permission string
		if err := rows.Scan(&role, &permission); err != nil {
			return nil, fmt.Errorf("database operation failed")
		}
		result[role] = append(result[role], permission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	return result, nil
}
