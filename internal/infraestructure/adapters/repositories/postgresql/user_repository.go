package postgresql

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

type UserSQLRepository struct {
	db *sql.DB
}

func (s *UserSQLRepository) handlePostgreSQLError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
			return &errors.ConflictError{Message: "user_already_exists"}
		}
	}
	return err
}

func (s *UserSQLRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// Extraer transacción del contexto si existe
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.createWithTx(ctx, tx, user)
	}

	// Si no hay transacción, usar conexión directa
	return s.createWithDB(ctx, user)
}

func (s *UserSQLRepository) createWithTx(ctx context.Context, tx *sql.Tx, user *models.User) (*models.User, error) {
	const query = `
		INSERT INTO users (name, last_name, email, password, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var userID int
	err := tx.QueryRowContext(ctx, query, user.Name, user.LastName, user.Email, user.Password, user.Phone).Scan(&userID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "create_user_tx",
			"error":     err.Error(),
		}).Error(DatabaseQueryFailedLog)
		return nil, s.handlePostgreSQLError(err)
	}

	user.ID = userID
	return user, nil
}

func (s *UserSQLRepository) createWithDB(ctx context.Context, user *models.User) (*models.User, error) {
	const query = `
		INSERT INTO users (name, last_name, email, password, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var userID int
	err := s.db.QueryRowContext(ctx, query, user.Name, user.LastName, user.Email, user.Password, user.Phone).Scan(&userID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "create_user_db",
			"error":     err.Error(),
		}).Error(DatabaseQueryFailedLog)
		return nil, s.handlePostgreSQLError(err)
	}

	user.ID = userID
	return user, nil
}

func (s *UserSQLRepository) AssignRole(ctx context.Context, userID int, roleID int) error {
	// Extraer transacción del contexto si existe
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.assignRoleWithTx(ctx, tx, userID, roleID)
	}

	// Si no hay transacción, usar conexión directa
	return s.assignRoleWithDB(ctx, userID, roleID)
}

func (s *UserSQLRepository) assignRoleWithTx(ctx context.Context, tx *sql.Tx, userID int, roleID int) error {
	const query = `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, now())
	`

	_, err := tx.ExecContext(ctx, query, userID, roleID)
	return err
}

func (s *UserSQLRepository) assignRoleWithDB(ctx context.Context, userID int, roleID int) error {
	const query = `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, now())
	`

	_, err := s.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (s *UserSQLRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	// Si hay transacción en contexto, úsala; sino conexión directa
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.getByEmailWithTx(ctx, tx, email)
	}
	return s.getByEmailWithDB(ctx, email)
}

func (s *UserSQLRepository) getByEmailWithTx(ctx context.Context, tx *sql.Tx, email string) (*models.User, error) {
	const query = `
		SELECT 
			u.id, u.name, u.email, u.phone, u.password, u.is_active,
			COALESCE(r.id, 0) as role_id, 
			COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		WHERE u.email = $1
		ORDER BY u.id, r.id`

	rows, err := tx.QueryContext(ctx, query, email)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "get_user_by_email_tx",
			"error":     err.Error(),
		}).Error(DatabaseQueryFailedLog)
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
	}
	defer rows.Close()

	return s.scanUserWithRoles(ctx, rows)
}

func (s *UserSQLRepository) getByEmailWithDB(ctx context.Context, email string) (*models.User, error) {
	const query = `
		SELECT 
			u.id, u.name, u.email, u.phone, u.password, u.is_active,
			COALESCE(r.id, 0) as role_id, 
			COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		WHERE u.email = $1
		ORDER BY u.id, r.id`

	rows, err := s.db.QueryContext(ctx, query, email)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "get_user_by_email_db",
			"error":     err.Error(),
		}).Error(DatabaseQueryFailedLog)
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
	}
	defer rows.Close()

	return s.scanUserWithRoles(ctx, rows)
}

func (s *UserSQLRepository) scanUserWithRoles(_ context.Context, rows *sql.Rows) (*models.User, error) {
	// Verificar si hay al menos una fila antes de procesar
	if !rows.Next() {
		// No hay datos, retornar directamente
		if err := rows.Err(); err != nil {
			return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
		}
		return nil, &errors.NotFoundError{Message: errors.UserNotFound}
	}

	// Hay datos, procesar la primera fila
	var user = &models.User{}
	var roles []*models.Role
	roleMap := make(map[int]bool) // Para evitar roles duplicados

	var roleID int
	var roleName string

	err := rows.Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.IsActive,
		&roleID, &roleName,
	)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "scan_user_row",
			"error":     err.Error(),
		}).Error(DatabaseScanFailedLog)
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
	}

	// Agregar el primer role si existe
	if roleID > 0 {
		roles = append(roles, &models.Role{
			ID:   roleID,
			Name: roleName,
		})
		roleMap[roleID] = true
	}

	// Procesar filas adicionales (roles adicionales)
	for rows.Next() {
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.IsActive,
			&roleID, &roleName,
		)
		if err != nil {
			logs.WithFields(map[string]interface{}{
				"operation": "scan_user_row_additional",
				"error":     err.Error(),
			}).Error(DatabaseScanFailedLog)
			return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
		}

		// Solo agregar role si existe y no está duplicado
		if roleID > 0 && !roleMap[roleID] {
			roles = append(roles, &models.Role{
				ID:   roleID,
				Name: roleName,
			})
			roleMap[roleID] = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
	}

	user.Roles = roles
	return user, nil
}

func NewUserRepository(dataBaseConnection DataBaseConnection) *UserSQLRepository {
	return &UserSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}
