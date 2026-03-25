package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const (
	StaffRepositoryField = "staff_repository"
)

type StaffSQLRepository struct {
	db *sql.DB
}

func NewStaffRepository(dataBaseConnection DataBaseConnection) ports.StaffRepository {
	return &StaffSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

// ==================== Internal methods (signup/signin) ====================

// Create creates a staff entry linking a user to a shop.
func (r *StaffSQLRepository) Create(ctx context.Context, userID, shopID int) (*models.Staff, error) {
	const query = `
		INSERT INTO staff (user_id, shop_id, is_active)
		VALUES ($1, $2, true)
		RETURNING id
	`

	var db interface {
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		db = tx
	} else {
		db = r.db
	}

	var staffID int
	err := db.QueryRowContext(ctx, query, userID, shopID).Scan(&staffID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PqErrCodeUniqueViolation {
			return nil, &errors.DuplicateRecordError{Message: errors.StaffAlreadyExists}
		}
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "create",
			"user_id":  userID,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to create staff")
		return nil, fmt.Errorf("database operation failed")
	}

	return &models.Staff{
		ID:       staffID,
		ShopID:   shopID,
		IsActive: true,
	}, nil
}

// AssignRole assigns a role to a staff member via staff_roles.
func (r *StaffSQLRepository) AssignRole(ctx context.Context, staffID, roleID int) error {
	const query = `
		INSERT INTO staff_roles (staff_id, role_id)
		VALUES ($1, $2)
	`

	var db interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		db = tx
	} else {
		db = r.db
	}

	_, err := db.ExecContext(ctx, query, staffID, roleID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "assign_role",
			"staff_id": staffID,
			"role_id":  roleID,
			"error":    err.Error(),
		}).Error("Failed to assign role to staff")
		return fmt.Errorf("database operation failed")
	}

	return nil
}

// GetShopRolesByUserID returns shop IDs with roles for a user.
// Used during sign-in to embed role-per-shop in the JWT token.
func (r *StaffSQLRepository) GetShopRolesByUserID(ctx context.Context, userID int) ([]claims.ShopRole, error) {
	const query = `
		SELECT s.shop_id, r.name as role
		FROM staff s
		JOIN staff_roles sr ON sr.staff_id = s.id
		JOIN roles r ON r.id = sr.role_id
		WHERE s.user_id = $1 AND s.is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "get_shop_roles_by_user_id",
			"user_id":  userID,
			"error":    err.Error(),
		}).Error("Failed to get shop roles by user ID")
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	shopRoles := make([]claims.ShopRole, 0)
	for rows.Next() {
		var sr claims.ShopRole
		if err := rows.Scan(&sr.ShopID, &sr.Role); err != nil {
			return nil, fmt.Errorf("database operation failed")
		}
		shopRoles = append(shopRoles, sr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	return shopRoles, nil
}

// ==================== CRUD methods (handler) ====================

// CreateWithUser creates a user, staff entry, staff_role, and optional image in a transaction.
func (r *StaffSQLRepository) CreateWithUser(ctx context.Context, shopID int, user *models.User, roleID int, profileImage *models.Image) (*models.Staff, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("database operation failed")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Create user
	userID, err := r.insertUser(ctx, tx, user)
	if err != nil {
		return nil, err
	}
	user.ID = userID

	// 2. Create staff entry
	newStaffID, err := r.insertStaff(ctx, tx, userID, shopID)
	if err != nil {
		return nil, err
	}

	// 3. Assign role
	if err = r.insertStaffRole(ctx, tx, newStaffID, roleID); err != nil {
		return nil, err
	}

	// 4. Insert profile image if provided
	if err = r.insertProfileImage(ctx, tx, profileImage, userID); err != nil {
		return nil, err
	}

	// 5. Commit
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	// 6. Get the role name for the response
	role := &models.Role{ID: roleID}
	if err := r.db.QueryRowContext(ctx, `SELECT name FROM roles WHERE id = $1`, roleID).Scan(&role.Name); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "create_with_user",
			"role_id":  roleID,
			"error":    err.Error(),
		}).Warn("Failed to fetch role name after commit")
	}

	return &models.Staff{
		ID:       newStaffID,
		ShopID:   shopID,
		IsActive: true,
		User:     user,
		Roles:    []*models.Role{role},
	}, nil
}

func (r *StaffSQLRepository) insertUser(ctx context.Context, tx *sql.Tx, user *models.User) (int, error) {
	var userID int
	err := tx.QueryRowContext(ctx,
		`INSERT INTO users (name, last_name, email, password, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		user.Name, user.LastName, user.Email, user.Password, user.IsActive,
	).Scan(&userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PqErrCodeUniqueViolation {
			return 0, &errors.DuplicateRecordError{Message: errors.UserAlreadyExists}
		}
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "create_with_user",
			"error":    err.Error(),
		}).Error("Failed to create user")
		return 0, fmt.Errorf("database operation failed")
	}
	return userID, nil
}

func (r *StaffSQLRepository) insertStaff(ctx context.Context, tx *sql.Tx, userID, shopID int) (int, error) {
	var staffID int
	err := tx.QueryRowContext(ctx,
		`INSERT INTO staff (user_id, shop_id, is_active)
		 VALUES ($1, $2, true)
		 RETURNING id`,
		userID, shopID,
	).Scan(&staffID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PqErrCodeUniqueViolation {
			return 0, &errors.DuplicateRecordError{Message: errors.StaffAlreadyExists}
		}
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "create_with_user",
			"error":    err.Error(),
		}).Error("Failed to create staff entry")
		return 0, fmt.Errorf("database operation failed")
	}
	return staffID, nil
}

func (r *StaffSQLRepository) insertStaffRole(ctx context.Context, tx *sql.Tx, staffID, roleID int) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO staff_roles (staff_id, role_id) VALUES ($1, $2)`,
		staffID, roleID,
	)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "create_with_user",
			"error":    err.Error(),
		}).Error("Failed to assign role to staff")
		return fmt.Errorf("database operation failed")
	}
	return nil
}

func (r *StaffSQLRepository) insertProfileImage(ctx context.Context, tx *sql.Tx, profileImage *models.Image, userID int) error {
	if profileImage == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO images (url, storage_ref, type, user_id)
		 VALUES ($1, $2, 'profile', $3)`,
		profileImage.URL, profileImage.StorageRef, userID,
	)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "create_with_user",
			"error":    err.Error(),
		}).Error("Failed to insert profile image")
		return fmt.Errorf("database operation failed")
	}
	return nil
}

const (
	staffDefaultSortColumn = "s.created_at"
	staffDefaultSortOrder  = "DESC"
)

// staffAllowedSortColumns whitelist for SQL injection defense-in-depth.
var staffAllowedSortColumns = map[string]string{
	"name":       "u.name",
	"created_at": "s.created_at",
}

var staffAllowedSortOrders = map[string]string{
	"asc":  "ASC",
	"desc": staffDefaultSortOrder,
}

// GetAllByShopIDWithFilters retrieves staff with filters and LIMIT+1 pagination.
func (r *StaffSQLRepository) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) ([]*models.Staff, error) {
	query := `
		SELECT
			s.id, s.is_active, s.created_at,
			u.id, u.name, u.last_name, u.email, u.is_active,
			COALESCE(
				(SELECT jsonb_agg(jsonb_build_object('id', rl.id, 'name', rl.name))
				 FROM staff_roles sr
				 JOIN roles rl ON rl.id = sr.role_id
				 WHERE sr.staff_id = s.id),
				'[]'::jsonb
			) AS roles,
			i.id AS image_id, i.url AS image_url
		FROM staff s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN images i ON i.user_id = u.id AND i.type = 'profile'
		WHERE s.shop_id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM staff_roles sr2
		      JOIN roles r2 ON r2.id = sr2.role_id
		      WHERE sr2.staff_id = s.id AND r2.name = 'owner'
		  )`

	args := []interface{}{shopID}

	// Search filter (name or email ILIKE)
	if filters.Search != nil && *filters.Search != "" {
		query += fmt.Sprintf(" AND (u.name ILIKE $%d OR u.last_name ILIKE $%d OR u.email ILIKE $%d)", len(args)+1, len(args)+1, len(args)+1)
		args = append(args, "%"+*filters.Search+"%")
	}

	// Active status filter (staff.is_active, not user.is_active)
	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND s.is_active = $%d", len(args)+1)
		args = append(args, *filters.IsActive)
	}

	// Defense-in-depth: resolve sort column and direction via whitelist
	sortColumn, ok := staffAllowedSortColumns[filters.SortBy]
	if !ok {
		sortColumn = staffDefaultSortColumn
	}
	sortDir, ok := staffAllowedSortOrders[filters.SortOrder]
	if !ok {
		sortDir = staffDefaultSortOrder
	}

	// Keyset pagination
	if filters.LastID != nil {
		condition, updatedArgs := r.buildStaffKeysetCondition(filters, sortColumn, args)
		query += condition
		args = updatedArgs
	}

	// Sort
	query += fmt.Sprintf(" ORDER BY %s %s, s.id %s", sortColumn, sortDir, sortDir)

	// LIMIT+1
	query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
	args = append(args, filters.Limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "get_all_by_shop_id_with_filters",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to get staff with filters")
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	staffList := make([]*models.Staff, 0)
	for rows.Next() {
		st := &models.Staff{ShopID: shopID}
		u := &models.User{}
		var rolesJSON []byte
		var imageID sql.NullInt64
		var imageURL sql.NullString

		err := rows.Scan(
			&st.ID, &st.IsActive, &st.CreatedAt,
			&u.ID, &u.Name, &u.LastName, &u.Email, &u.IsActive,
			&rolesJSON,
			&imageID, &imageURL,
		)
		if err != nil {
			return nil, fmt.Errorf("database operation failed")
		}

		if imageID.Valid {
			u.ProfileImage = &models.Image{
				ID:  int(imageID.Int64),
				URL: imageURL.String,
			}
		}

		var roles []*models.Role
		if err := json.Unmarshal(rolesJSON, &roles); err != nil {
			return nil, fmt.Errorf("database operation failed")
		}

		st.User = u
		st.Roles = roles
		staffList = append(staffList, st)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	return staffList, nil
}

// buildStaffKeysetCondition builds the WHERE clause fragment for keyset pagination.
func (r *StaffSQLRepository) buildStaffKeysetCondition(filters models.StaffFilters, sortColumn string, args []interface{}) (string, []interface{}) {
	isDesc := filters.SortOrder == models.SortOrderDesc

	op := ">"
	if isDesc {
		op = "<"
	}

	if filters.LastSortValue != nil {
		condition := fmt.Sprintf(
			" AND (%s %s $%d OR (%s = $%d AND s.id %s $%d))",
			sortColumn, op, len(args)+1, sortColumn, len(args)+1, op, len(args)+2,
		)
		args = append(args, filters.LastSortValue, *filters.LastID)
		return condition, args
	}

	condition := fmt.Sprintf(" AND s.id %s $%d", op, len(args)+1)
	args = append(args, *filters.LastID)
	return condition, args
}

// CountByShopIDWithFilters returns total count of staff matching filters.
func (r *StaffSQLRepository) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) (int, error) {
	query := `SELECT COUNT(*) FROM staff s JOIN users u ON u.id = s.user_id
		WHERE s.shop_id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM staff_roles sr2
		      JOIN roles r2 ON r2.id = sr2.role_id
		      WHERE sr2.staff_id = s.id AND r2.name = 'owner'
		  )`
	args := []interface{}{shopID}

	if filters.Search != nil && *filters.Search != "" {
		query += fmt.Sprintf(" AND (u.name ILIKE $%d OR u.last_name ILIKE $%d OR u.email ILIKE $%d)", len(args)+1, len(args)+1, len(args)+1)
		args = append(args, "%"+*filters.Search+"%")
	}

	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND s.is_active = $%d", len(args)+1)
		args = append(args, *filters.IsActive)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "count_by_shop_id_with_filters",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to count staff")
		return 0, fmt.Errorf("database operation failed")
	}

	return count, nil
}

// GetByID returns a single staff member by ID and shop ID.
func (r *StaffSQLRepository) GetByID(ctx context.Context, staffID, shopID int) (*models.Staff, error) {
	const query = `
		SELECT
			s.id, s.is_active,
			u.id, u.name, u.last_name, u.email, u.is_active,
			COALESCE(
				(SELECT jsonb_agg(jsonb_build_object('id', rl.id, 'name', rl.name))
				 FROM staff_roles sr
				 JOIN roles rl ON rl.id = sr.role_id
				 WHERE sr.staff_id = s.id),
				'[]'::jsonb
			) AS roles,
			i.id AS image_id, i.url AS image_url, i.storage_ref AS image_storage_ref
		FROM staff s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN images i ON i.user_id = u.id AND i.type = 'profile'
		WHERE s.id = $1 AND s.shop_id = $2
	`

	st := &models.Staff{ShopID: shopID}
	u := &models.User{}
	var rolesJSON []byte
	var imageID sql.NullInt64
	var imageURL sql.NullString
	var imageStorageRef sql.NullString

	err := r.db.QueryRowContext(ctx, query, staffID, shopID).Scan(
		&st.ID, &st.IsActive,
		&u.ID, &u.Name, &u.LastName, &u.Email, &u.IsActive,
		&rolesJSON,
		&imageID, &imageURL, &imageStorageRef,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errors.RecordNotFoundError{Message: errors.StaffNotFound}
		}
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "get_by_id",
			"staff_id": staffID,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to get staff")
		return nil, fmt.Errorf("database operation failed")
	}

	if imageID.Valid {
		u.ProfileImage = &models.Image{
			ID:         int(imageID.Int64),
			URL:        imageURL.String,
			StorageRef: imageStorageRef.String,
		}
	}

	var roles []*models.Role
	if err := json.Unmarshal(rolesJSON, &roles); err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	st.User = u
	st.Roles = roles
	return st, nil
}

// Update updates a staff member's user data, role, and optional image.
func (r *StaffSQLRepository) Update(ctx context.Context, staffID, shopID int, user *models.User, roleID int, profileImage *models.Image) (*models.Staff, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("database operation failed")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Get user_id from staff
	var userID int
	err = tx.QueryRowContext(ctx,
		`SELECT user_id FROM staff WHERE id = $1 AND shop_id = $2`,
		staffID, shopID,
	).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errors.RecordNotFoundError{Message: errors.StaffNotFound}
		}
		return nil, fmt.Errorf("database operation failed")
	}

	// 2. Update user data
	if user.Password != "" {
		_, err = tx.ExecContext(ctx,
			`UPDATE users SET name = $1, last_name = $2, email = $3, password = $4 WHERE id = $5`,
			user.Name, user.LastName, user.Email, user.Password, userID,
		)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE users SET name = $1, last_name = $2, email = $3 WHERE id = $4`,
			user.Name, user.LastName, user.Email, userID,
		)
	}
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PqErrCodeUniqueViolation {
			return nil, &errors.DuplicateRecordError{Message: errors.UserAlreadyExists}
		}
		return nil, fmt.Errorf("database operation failed")
	}

	// 3. Update role: delete existing, insert new
	_, err = tx.ExecContext(ctx, `DELETE FROM staff_roles WHERE staff_id = $1`, staffID)
	if err != nil {
		return nil, fmt.Errorf("database operation failed")
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO staff_roles (staff_id, role_id) VALUES ($1, $2)`,
		staffID, roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	// 4. Update profile image if provided
	if profileImage != nil {
		// Delete old image
		_, _ = tx.ExecContext(ctx, `DELETE FROM images WHERE user_id = $1 AND type = 'profile'`, userID)
		// Insert new image
		_, err = tx.ExecContext(ctx,
			`INSERT INTO images (url, storage_ref, type, user_id) VALUES ($1, $2, 'profile', $3)`,
			profileImage.URL, profileImage.StorageRef, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("database operation failed")
		}
	}

	// 5. Commit
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	// 6. Return fresh data
	return r.GetByID(ctx, staffID, shopID)
}

// Delete removes a staff member (cascade deletes staff_roles).
func (r *StaffSQLRepository) Delete(ctx context.Context, staffID, shopID int) error {
	const query = `DELETE FROM staff WHERE id = $1 AND shop_id = $2`

	result, err := r.db.ExecContext(ctx, query, staffID, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "delete",
			"staff_id": staffID,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to delete staff")
		return fmt.Errorf("database operation failed")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("database operation failed")
	}

	if rowsAffected == 0 {
		return &errors.RecordNotFoundError{Message: errors.StaffNotFound}
	}

	return nil
}

// ToggleStatus toggles the is_active flag on a staff member.
func (r *StaffSQLRepository) ToggleStatus(ctx context.Context, staffID, shopID int) (*models.Staff, error) {
	const query = `
		UPDATE staff SET is_active = NOT is_active
		WHERE id = $1 AND shop_id = $2
		RETURNING is_active
	`

	var isActive bool
	err := r.db.QueryRowContext(ctx, query, staffID, shopID).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errors.RecordNotFoundError{Message: errors.StaffNotFound}
		}
		logs.WithFields(map[string]interface{}{
			"file":     StaffRepositoryField,
			"function": "toggle_status",
			"staff_id": staffID,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to toggle staff status")
		return nil, fmt.Errorf("database operation failed")
	}

	// Return fresh data with all relations
	return r.GetByID(ctx, staffID, shopID)
}
