package postgresql

import (
	"context"
	"database/sql"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type RoleSQLRepository struct {
	db *sql.DB
}

func NewRoleRepository(dataBaseConnection DataBaseConnection) ports.RoleRepository {
	return &RoleSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

func (r *RoleSQLRepository) GetByID(ctx context.Context, id int) (*models.Role, error) {
	const query = `SELECT id, name, description FROM roles WHERE id = $1`

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, id).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &role, nil
}

func (r *RoleSQLRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	// Extraer transacción del contexto si existe
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.getByNameWithTx(ctx, tx, name)
	}

	// Si no hay transacción, usar conexión directa
	return r.getByNameWithDB(ctx, name)
}

func (r *RoleSQLRepository) getByNameWithTx(ctx context.Context, tx *sql.Tx, name string) (*models.Role, error) {
	const query = `SELECT id, name, description FROM roles WHERE name = $1`

	var role models.Role
	err := tx.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleSQLRepository) getByNameWithDB(ctx context.Context, name string) (*models.Role, error) {
	const query = `SELECT id, name, description FROM roles WHERE name = $1`

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleSQLRepository) GetAllAssignable(ctx context.Context) ([]*models.Role, error) {
	const query = `SELECT id, name, description FROM roles WHERE name != 'owner' ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}
