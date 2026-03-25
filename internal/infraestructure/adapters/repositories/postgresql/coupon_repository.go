package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const (
	CouponRepositoryField       = "coupon_repository"
	CouponCreateFuncField       = "create"
	CouponUpdateFuncField       = "update"
	CouponDeleteFuncField       = "delete"
	CouponGetByIDFuncField      = "get_by_id"
	CouponGetByCodeFuncField    = "get_by_code_and_shop_id"
	CouponGetAllFuncField       = "get_all_by_shop_id_with_filters"
	CouponCountFuncField        = "count_by_shop_id_with_filters"
	CouponCountUsagesFuncField  = "count_usages"
	CouponCountByPhoneFuncField = "count_usages_by_phone"

	ConstraintCouponCodeShopIDUnique = "coupons_code_shop_id_unique"
)

// couponAllowedSortColumns whitelist for SQL injection defense-in-depth.
// Even though CouponFilters.Validated() checks these, the repository must
// independently guarantee only safe column names reach the query.
var couponAllowedSortColumns = map[string]string{
	"code":       "c.code",
	"created_at": "c.created_at",
}

const couponDefaultSortOrder = "DESC"

// couponAllowedSortOrders whitelist for ORDER BY direction.
var couponAllowedSortOrders = map[string]string{
	"asc":  "ASC",
	"desc": couponDefaultSortOrder,
}

type CouponSQLRepository struct {
	db *sql.DB
}

func NewCouponRepository(dataBaseConnection DataBaseConnection) ports.CouponRepository {
	return &CouponSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

// Create inserts a new coupon into the database.
func (r *CouponSQLRepository) Create(ctx context.Context, coupon *models.Coupon) (*models.Coupon, error) {
	query := `
		INSERT INTO coupons (shop_id, code, type, value, min_order_amount,
			usage_limit, max_uses_per_phone, starts_at, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		coupon.ShopID,
		coupon.Code,
		coupon.Type,
		coupon.Value,
		nullFloat64(coupon.MinOrderAmount),
		coupon.UsageLimit,
		coupon.MaxUsesPerPhone,
		coupon.StartsAt,
		coupon.ExpiresAt,
		coupon.IsActive,
	).Scan(&coupon.ID, &coupon.CreatedAt, &coupon.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == PqErrCodeUniqueViolation && pqErr.Constraint == ConstraintCouponCodeShopIDUnique {
				return nil, &errors.DuplicateRecordError{Message: errors.CouponAlreadyExists}
			}
		}
		logs.WithFields(map[string]interface{}{
			"file":     CouponRepositoryField,
			"function": CouponCreateFuncField,
			"shop_id":  coupon.ShopID,
			"code":     coupon.Code,
			"error":    err.Error(),
		}).Error("Failed to create coupon")
		return nil, fmt.Errorf("database operation failed")
	}

	return coupon, nil
}

// Update updates an existing coupon.
func (r *CouponSQLRepository) Update(ctx context.Context, coupon *models.Coupon) error {
	query := `
		UPDATE coupons
		SET code = $1, type = $2, value = $3, min_order_amount = $4,
			usage_limit = $5, max_uses_per_phone = $6, starts_at = $7,
			expires_at = $8, is_active = $9, updated_at = now()
		WHERE id = $10 AND shop_id = $11`

	result, err := r.db.ExecContext(ctx, query,
		coupon.Code,
		coupon.Type,
		coupon.Value,
		nullFloat64(coupon.MinOrderAmount),
		coupon.UsageLimit,
		coupon.MaxUsesPerPhone,
		coupon.StartsAt,
		coupon.ExpiresAt,
		coupon.IsActive,
		coupon.ID,
		coupon.ShopID,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == PqErrCodeUniqueViolation && pqErr.Constraint == ConstraintCouponCodeShopIDUnique {
				return &errors.DuplicateRecordError{Message: errors.CouponAlreadyExists}
			}
		}
		logs.WithFields(map[string]interface{}{
			"file":      CouponRepositoryField,
			"function":  CouponUpdateFuncField,
			"coupon_id": coupon.ID,
			"shop_id":   coupon.ShopID,
			"error":     err.Error(),
		}).Error("Failed to update coupon")
		return fmt.Errorf("database operation failed")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("database operation failed")
	}
	if rowsAffected == 0 {
		return &errors.RecordNotFoundError{Message: errors.CouponNotFound}
	}

	return nil
}

// Delete deletes a coupon by ID and shop ID.
func (r *CouponSQLRepository) Delete(ctx context.Context, couponID, shopID int) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM coupons WHERE id = $1 AND shop_id = $2",
		couponID, shopID,
	)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      CouponRepositoryField,
			"function":  CouponDeleteFuncField,
			"coupon_id": couponID,
			"shop_id":   shopID,
			"error":     err.Error(),
		}).Error("Failed to delete coupon")
		return fmt.Errorf("database operation failed")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("database operation failed")
	}
	if rowsAffected == 0 {
		return &errors.RecordNotFoundError{Message: errors.CouponNotFound}
	}

	return nil
}

// GetByID retrieves a coupon by ID and shop ID, including usage count.
func (r *CouponSQLRepository) GetByID(ctx context.Context, couponID, shopID int) (*models.Coupon, error) {
	query := `
		SELECT c.id, c.shop_id, c.code, c.type, c.value, c.min_order_amount,
			   c.usage_limit, c.max_uses_per_phone, c.starts_at, c.expires_at,
			   c.is_active, c.created_at, c.updated_at,
			   COUNT(cu.id) AS usage_count
		FROM coupons c
		LEFT JOIN coupon_usages cu ON cu.coupon_id = c.id
		WHERE c.id = $1 AND c.shop_id = $2
		GROUP BY c.id`

	coupon, err := r.scanCouponWithUsageCount(r.db.QueryRowContext(ctx, query, couponID, shopID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errors.RecordNotFoundError{Message: errors.CouponNotFound}
		}
		logs.WithFields(map[string]interface{}{
			"file":      CouponRepositoryField,
			"function":  CouponGetByIDFuncField,
			"coupon_id": couponID,
			"shop_id":   shopID,
			"error":     err.Error(),
		}).Error("Failed to get coupon by ID")
		return nil, fmt.Errorf("database operation failed")
	}

	return coupon, nil
}

// GetByCodeAndShopID retrieves a coupon by code and shop ID.
func (r *CouponSQLRepository) GetByCodeAndShopID(ctx context.Context, code string, shopID int) (*models.Coupon, error) {
	query := `
		SELECT id, shop_id, code, type, value, min_order_amount,
			   usage_limit, max_uses_per_phone, starts_at, expires_at,
			   is_active, created_at, updated_at
		FROM coupons
		WHERE code = $1 AND shop_id = $2`

	coupon, err := r.scanCoupon(r.db.QueryRowContext(ctx, query, code, shopID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errors.RecordNotFoundError{Message: errors.CouponNotFound}
		}
		logs.WithFields(map[string]interface{}{
			"file":     CouponRepositoryField,
			"function": CouponGetByCodeFuncField,
			"code":     code,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to get coupon by code")
		return nil, fmt.Errorf("database operation failed")
	}

	return coupon, nil
}

// GetAllByShopIDWithFilters retrieves coupons with filters and LIMIT+1 pagination.
func (r *CouponSQLRepository) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) ([]*models.Coupon, error) {
	query := `
		SELECT c.id, c.shop_id, c.code, c.type, c.value, c.min_order_amount,
			   c.usage_limit, c.max_uses_per_phone, c.starts_at, c.expires_at,
			   c.is_active, c.created_at, c.updated_at,
			   COUNT(cu.id) AS usage_count
		FROM coupons c
		LEFT JOIN coupon_usages cu ON cu.coupon_id = c.id
		WHERE c.shop_id = $1`

	args := []interface{}{shopID}

	// Search filter (code ILIKE)
	if filters.Search != nil && *filters.Search != "" {
		query += fmt.Sprintf(" AND c.code ILIKE $%d", len(args)+1)
		args = append(args, "%"+*filters.Search+"%")
	}

	// Active status filter
	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND c.is_active = $%d", len(args)+1)
		args = append(args, *filters.IsActive)
	}

	// Defense-in-depth: resolve sort column and direction via whitelist.
	// Even if CouponFilters.Validated() already checked these values,
	// the repository guarantees only safe identifiers reach the query.
	sortColumn, ok := couponAllowedSortColumns[filters.SortBy]
	if !ok {
		sortColumn = "c.created_at"
	}
	sortDir, ok := couponAllowedSortOrders[filters.SortOrder]
	if !ok {
		sortDir = couponDefaultSortOrder
	}

	// Keyset pagination: use (sort_field, id) composite cursor
	// This ensures stable pagination even when sort_field has duplicate values
	if filters.LastID != nil {
		condition, updatedArgs := r.buildKeysetCondition(filters, sortColumn, args)
		query += condition
		args = updatedArgs
	}

	// GROUP BY before ORDER BY
	query += " GROUP BY c.id"

	// Sort (using whitelisted column and direction)
	query += fmt.Sprintf(" ORDER BY %s %s, c.id %s", sortColumn, sortDir, sortDir)

	// LIMIT+1
	query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
	args = append(args, filters.Limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CouponRepositoryField,
			"function": CouponGetAllFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to get coupons with filters")
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	coupons := make([]*models.Coupon, 0)
	for rows.Next() {
		coupon, err := r.scanCouponRowWithUsageCount(rows)
		if err != nil {
			return nil, fmt.Errorf("database operation failed")
		}
		coupons = append(coupons, coupon)
	}

	return coupons, rows.Err()
}

// buildKeysetCondition builds the WHERE clause fragment for keyset pagination
// using a composite cursor of (sort_field, id). Returns the SQL condition and
// updated args slice.
func (r *CouponSQLRepository) buildKeysetCondition(filters models.CouponFilters, sortColumn string, args []interface{}) (string, []interface{}) {
	isDesc := filters.SortOrder == models.SortOrderDesc

	op := ">"
	if isDesc {
		op = "<"
	}

	if filters.LastSortValue != nil {
		condition := fmt.Sprintf(
			" AND (%s %s $%d OR (%s = $%d AND c.id %s $%d))",
			sortColumn, op, len(args)+1, sortColumn, len(args)+1, op, len(args)+2,
		)
		args = append(args, filters.LastSortValue, *filters.LastID)
		return condition, args
	}

	condition := fmt.Sprintf(" AND c.id %s $%d", op, len(args)+1)
	args = append(args, *filters.LastID)
	return condition, args
}

// CountByShopIDWithFilters returns total count of coupons matching filters.
func (r *CouponSQLRepository) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) (int, error) {
	query := "SELECT COUNT(*) FROM coupons WHERE shop_id = $1"
	args := []interface{}{shopID}

	if filters.Search != nil && *filters.Search != "" {
		query += fmt.Sprintf(" AND code ILIKE $%d", len(args)+1)
		args = append(args, "%"+*filters.Search+"%")
	}

	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *filters.IsActive)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CouponRepositoryField,
			"function": CouponCountFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to count coupons")
		return 0, fmt.Errorf("database operation failed")
	}

	return count, nil
}

// CountUsages returns the total number of times a coupon has been used.
func (r *CouponSQLRepository) CountUsages(ctx context.Context, couponID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM coupon_usages WHERE coupon_id = $1",
		couponID,
	).Scan(&count)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      CouponRepositoryField,
			"function":  CouponCountUsagesFuncField,
			"coupon_id": couponID,
			"error":     err.Error(),
		}).Error("Failed to count coupon usages")
		return 0, fmt.Errorf("database operation failed")
	}
	return count, nil
}

// CountUsagesByPhone returns the number of times a coupon has been used by a specific phone.
func (r *CouponSQLRepository) CountUsagesByPhone(ctx context.Context, couponID int, phone string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM coupon_usages WHERE coupon_id = $1 AND phone = $2",
		couponID, phone,
	).Scan(&count)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      CouponRepositoryField,
			"function":  CouponCountByPhoneFuncField,
			"coupon_id": couponID,
			"phone":     phone,
			"error":     err.Error(),
		}).Error("Failed to count coupon usages by phone")
		return 0, fmt.Errorf("database operation failed")
	}
	return count, nil
}

// scanCoupon scans a single coupon row (without usage count).
func (r *CouponSQLRepository) scanCoupon(row *sql.Row) (*models.Coupon, error) {
	var coupon models.Coupon
	var minOrderAmount sql.NullFloat64
	var usageLimit, maxUsesPerPhone sql.NullInt32
	var startsAt, expiresAt sql.NullTime

	err := row.Scan(
		&coupon.ID,
		&coupon.ShopID,
		&coupon.Code,
		&coupon.Type,
		&coupon.Value,
		&minOrderAmount,
		&usageLimit,
		&maxUsesPerPhone,
		&startsAt,
		&expiresAt,
		&coupon.IsActive,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	applyNullableCouponFields(&coupon, minOrderAmount, usageLimit, maxUsesPerPhone, startsAt, expiresAt)
	return &coupon, nil
}

// scanCouponWithUsageCount scans a single coupon row including usage count.
func (r *CouponSQLRepository) scanCouponWithUsageCount(row *sql.Row) (*models.Coupon, error) {
	var coupon models.Coupon
	var minOrderAmount sql.NullFloat64
	var usageLimit, maxUsesPerPhone sql.NullInt32
	var startsAt, expiresAt sql.NullTime

	err := row.Scan(
		&coupon.ID,
		&coupon.ShopID,
		&coupon.Code,
		&coupon.Type,
		&coupon.Value,
		&minOrderAmount,
		&usageLimit,
		&maxUsesPerPhone,
		&startsAt,
		&expiresAt,
		&coupon.IsActive,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
		&coupon.UsageCount,
	)
	if err != nil {
		return nil, err
	}

	applyNullableCouponFields(&coupon, minOrderAmount, usageLimit, maxUsesPerPhone, startsAt, expiresAt)
	return &coupon, nil
}

// scanCouponRowWithUsageCount scans a coupon from rows (query result set) including usage count.
func (r *CouponSQLRepository) scanCouponRowWithUsageCount(rows *sql.Rows) (*models.Coupon, error) {
	var coupon models.Coupon
	var minOrderAmount sql.NullFloat64
	var usageLimit, maxUsesPerPhone sql.NullInt32
	var startsAt, expiresAt sql.NullTime

	err := rows.Scan(
		&coupon.ID,
		&coupon.ShopID,
		&coupon.Code,
		&coupon.Type,
		&coupon.Value,
		&minOrderAmount,
		&usageLimit,
		&maxUsesPerPhone,
		&startsAt,
		&expiresAt,
		&coupon.IsActive,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
		&coupon.UsageCount,
	)
	if err != nil {
		return nil, err
	}

	applyNullableCouponFields(&coupon, minOrderAmount, usageLimit, maxUsesPerPhone, startsAt, expiresAt)
	return &coupon, nil
}

// applyNullableCouponFields maps SQL nullable values to coupon model fields.
func applyNullableCouponFields(
	coupon *models.Coupon,
	minOrderAmount sql.NullFloat64,
	usageLimit, maxUsesPerPhone sql.NullInt32,
	startsAt, expiresAt sql.NullTime,
) {
	if minOrderAmount.Valid {
		coupon.MinOrderAmount = minOrderAmount.Float64
	}
	if usageLimit.Valid {
		v := int(usageLimit.Int32)
		coupon.UsageLimit = &v
	}
	if maxUsesPerPhone.Valid {
		v := int(maxUsesPerPhone.Int32)
		coupon.MaxUsesPerPhone = &v
	}
	if startsAt.Valid {
		coupon.StartsAt = &startsAt.Time
	}
	if expiresAt.Valid {
		coupon.ExpiresAt = &expiresAt.Time
	}
}

// nullFloat64 returns nil for zero values, allowing SQL NULL insertion.
// For min_order_amount, zero means "no minimum" — semantically equivalent to NULL.
func nullFloat64(v float64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
