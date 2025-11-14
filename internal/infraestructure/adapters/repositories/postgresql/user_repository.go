package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// User repository log field constants
const (
	UserRepositoryField           = "user_repository"
	UserCreateFunctionField       = "create"
	UserGetByEmailFunctionField   = "get_by_email"
	UserAssignRoleFunctionField   = "assign_role"
	UserScanWithRolesSubFuncField = "scan_user_with_roles"
)

type UserSQLRepository struct {
	db *sql.DB
}

// handlePostgreSQLError translates PostgreSQL errors to domain errors
func (s *UserSQLRepository) handlePostgreSQLError(err error, email string) error {
	if pqErr, ok := err.(*pq.Error); ok {
		// Duplicate email (unique constraint violation)
		if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
			logs.WithFields(map[string]interface{}{
				"file":       UserRepositoryField,
				"function":   UserCreateFunctionField,
				"constraint": pqErr.Constraint,
				"email":      email,
			}).Error("User with email already exists")

			return &errors.DuplicateRecordError{
				Message: errors.UserAlreadyExists,
			}
		}
	}
	// Technical error - log details but return generic error
	logs.WithFields(map[string]interface{}{
		"file":     UserRepositoryField,
		"function": UserCreateFunctionField,
		"error":    err.Error(),
	}).Error("Database error creating user")

	return fmt.Errorf("failed to create user")
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
		return nil, s.handlePostgreSQLError(err, user.Email)
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
		return nil, s.handlePostgreSQLError(err, user.Email)
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
			"file":     UserRepositoryField,
			"function": UserGetByEmailFunctionField,
			"sub_func": "tx.QueryContext",
			"email":    email,
			"error":    err.Error(),
		}).Error("Database query failed")
		return nil, fmt.Errorf("failed to get user by email")
	}
	defer rows.Close()

	return s.scanUserWithRoles(ctx, rows, email)
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
			"file":     UserRepositoryField,
			"function": UserGetByEmailFunctionField,
			"sub_func": "db.QueryContext",
			"email":    email,
			"error":    err.Error(),
		}).Error("Database query failed")
		return nil, fmt.Errorf("failed to get user by email")
	}
	defer rows.Close()

	return s.scanUserWithRoles(ctx, rows, email)
}

func (s *UserSQLRepository) scanUserWithRoles(_ context.Context, rows *sql.Rows, email string) (*models.User, error) {
	// Verificar si hay al menos una fila antes de procesar
	if !rows.Next() {
		// No hay datos - usuario no encontrado
		if err := rows.Err(); err != nil {
			logs.WithFields(map[string]interface{}{
				"file":     UserRepositoryField,
				"function": UserScanWithRolesSubFuncField,
				"sub_func": "rows.Next",
				"email":    email,
				"error":    err.Error(),
			}).Error("Database scan failed")
			return nil, fmt.Errorf("failed to scan user rows")
		}

		// Domain error: user not found
		logs.WithFields(map[string]interface{}{
			"file":     UserRepositoryField,
			"function": UserScanWithRolesSubFuncField,
			"email":    email,
		}).Error("User not found")

		return nil, &errors.RecordNotFoundError{
			Message: errors.UserNotFound,
		}
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
			"file":     UserRepositoryField,
			"function": UserScanWithRolesSubFuncField,
			"sub_func": "rows.Scan",
			"email":    email,
			"error":    err.Error(),
		}).Error("Database scan failed")
		return nil, fmt.Errorf("failed to scan user row")
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
				"file":     UserRepositoryField,
				"function": UserScanWithRolesSubFuncField,
				"sub_func": "rows.Scan",
				"email":    email,
				"error":    err.Error(),
			}).Error("Database scan failed on additional roles")
			return nil, fmt.Errorf("failed to scan user roles")
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
		logs.WithFields(map[string]interface{}{
			"file":     UserRepositoryField,
			"function": UserScanWithRolesSubFuncField,
			"sub_func": "rows.Err",
			"email":    email,
			"error":    err.Error(),
		}).Error("Database rows iteration error")
		return nil, fmt.Errorf("failed to iterate user rows")
	}

	user.Roles = roles
	return user, nil
}

func NewUserRepository(dataBaseConnection DataBaseConnection) *UserSQLRepository {
	return &UserSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}
