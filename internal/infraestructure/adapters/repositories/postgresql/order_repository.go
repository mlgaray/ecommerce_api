package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Order repository log field constants
const (
	OrderRepositoryField                    = "order_repository"
	OrderCreateFunctionField                = "create"
	OrderGetAllByShopIDWithFiltersFuncField = "get_all_by_shop_id_with_filters"
	OrderCountByShopIDWithFiltersFuncField  = "count_by_shop_id_with_filters"
)

type OrderSQLRepository struct {
	db *sql.DB
}

func NewOrderRepository(dataBaseConnection DataBaseConnection) ports.OrderRepository {
	return &OrderSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

// Create creates a new order with all its items and selected options.
// Uses stored procedure to handle transactional insert.
func (r *OrderSQLRepository) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	startTime := time.Now()

	// 1. Serialize items to JSONB
	itemsJSON, err := r.serializeItems(order.Items)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderCreateFunctionField,
			"sub_func": "serialize_items",
			"error":    err.Error(),
		}).Error("Failed to serialize order items")
		return nil, fmt.Errorf("database operation failed")
	}

	// 2. Extract address fields (can be nil)
	var addressName, addressPlaceID *string
	var addressLat, addressLng *float64
	if order.Customer != nil && order.Customer.Address != nil {
		addressName = &order.Customer.Address.Name
		addressPlaceID = &order.Customer.Address.PlaceID
		addressLat = &order.Customer.Address.Lat
		addressLng = &order.Customer.Address.Lng
	}

	// 3. Extract payment method fields (can be nil)
	var paymentMethodID *int
	var paymentMethodCode, paymentMethodName *string
	if order.PaymentMethod != nil {
		paymentMethodID = &order.PaymentMethod.ID
		code := string(order.PaymentMethod.Code)
		paymentMethodCode = &code
		paymentMethodName = &order.PaymentMethod.Name
	}

	// 4. Extract delivery method fields (can be nil)
	var deliveryMethodID *int
	var deliveryMethodCode, deliveryMethodName *string
	if order.DeliveryMethod != nil {
		deliveryMethodID = &order.DeliveryMethod.ID
		code := string(order.DeliveryMethod.Code)
		deliveryMethodCode = &code
		deliveryMethodName = &order.DeliveryMethod.Name
	}

	// 5. Extract customer fields
	var customerName, customerPhone, customerEmail *string
	if order.Customer != nil {
		customerName = &order.Customer.Name
		customerPhone = &order.Customer.Phone
		customerEmail = &order.Customer.Email
	}

	// 6. Call stored procedure
	var resultJSON []byte
	err = r.db.QueryRowContext(ctx, `
		SELECT create_order(
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		)`,
		order.Store.ID,
		order.Store.Name,
		order.Store.Slug,
		customerName,
		customerPhone,
		customerEmail,
		addressName,
		addressPlaceID,
		addressLat,
		addressLng,
		paymentMethodID,
		paymentMethodCode,
		paymentMethodName,
		deliveryMethodID,
		deliveryMethodCode,
		deliveryMethodName,
		order.ShippingCost,
		itemsJSON,
	).Scan(&resultJSON)

	if err != nil {
		return r.handleCreateError(err, order.Store.ID)
	}

	// 7. Parse result
	var result struct {
		ID           int       `json:"id"`
		OrderNumber  string    `json:"order_number"`
		Status       string    `json:"status"`
		Subtotal     float64   `json:"subtotal"`
		ShippingCost float64   `json:"shipping_cost"`
		Total        float64   `json:"total"`
		CreatedAt    time.Time `json:"created_at"`
	}

	if err := json.Unmarshal(resultJSON, &result); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderCreateFunctionField,
			"sub_func": UnmarshallField,
			"error":    err.Error(),
		}).Error("Failed to unmarshal stored procedure result")
		return nil, fmt.Errorf("database operation failed")
	}

	// 8. Update order with DB-generated values
	order.ID = result.ID
	order.OrderNumber = result.OrderNumber
	order.Status = models.OrderStatus(result.Status)
	order.Subtotal = result.Subtotal
	order.ShippingCost = result.ShippingCost
	order.Total = result.Total
	order.CreatedAt = result.CreatedAt

	logs.WithFields(map[string]interface{}{
		"file":              OrderRepositoryField,
		"function":          OrderCreateFunctionField,
		"order_id":          order.ID,
		"order_number":      order.OrderNumber,
		"store_id":          order.Store.ID,
		"items_count":       len(order.Items),
		"total_duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Order creation completed (stored procedure)")

	return order, nil
}

// serializeItems converts order items to JSONB format for the stored procedure.
func (r *OrderSQLRepository) serializeItems(items []*models.OrderItem) ([]byte, error) {
	type selectedOptionJSON struct {
		VariantID   int     `json:"variant_id"`
		OptionID    int     `json:"option_id"`
		VariantName string  `json:"variant_name"`
		OptionName  string  `json:"option_name"`
		OptionPrice float64 `json:"option_price"`
	}

	type itemJSON struct {
		ProductID        int                  `json:"product_id"`
		ProductName      string               `json:"product_name"`
		ProductImageURL  string               `json:"product_image_url,omitempty"`
		BasePrice        float64              `json:"base_price"`
		IsPromotional    bool                 `json:"is_promotional"`
		PromotionalPrice float64              `json:"promotional_price,omitempty"`
		Quantity         int                  `json:"quantity"`
		UnitPrice        float64              `json:"unit_price"`
		TotalPrice       float64              `json:"total_price"`
		SelectedOptions  []selectedOptionJSON `json:"selected_options,omitempty"`
	}

	jsonItems := make([]itemJSON, 0, len(items))
	for _, item := range items {
		jItem := itemJSON{
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		}

		if item.Product != nil {
			jItem.ProductID = item.Product.ID
			jItem.ProductName = item.Product.Name
			jItem.BasePrice = item.Product.Price
			jItem.IsPromotional = item.Product.IsPromotional
			jItem.PromotionalPrice = item.Product.PromotionalPrice

			// Get first image URL if available
			if len(item.Product.Images) > 0 {
				jItem.ProductImageURL = item.Product.Images[0].URL
			}

			// Serialize selected options (from variants)
			for _, variant := range item.Product.Variants {
				for _, option := range variant.Options {
					jItem.SelectedOptions = append(jItem.SelectedOptions, selectedOptionJSON{
						VariantID:   variant.ID,
						OptionID:    option.ID,
						VariantName: variant.Name,
						OptionName:  option.Name,
						OptionPrice: option.Price,
					})
				}
			}
		}

		jsonItems = append(jsonItems, jItem)
	}

	return json.Marshal(jsonItems)
}

// handleCreateError processes PostgreSQL errors from the stored procedure.
func (r *OrderSQLRepository) handleCreateError(err error, storeID int) (*models.Order, error) {
	logs.WithFields(map[string]interface{}{
		"file":     OrderRepositoryField,
		"function": OrderCreateFunctionField,
		"store_id": storeID,
		"error":    err.Error(),
	}).Error("Failed to create order via stored procedure")

	// Check if it's a PostgreSQL error
	if pqErr, ok := err.(*pq.Error); ok {
		// Store not found (P0002)
		if pqErr.Code == PqErrCodeNoDataFound {
			logs.WithFields(map[string]interface{}{
				"file":     OrderRepositoryField,
				"function": OrderCreateFunctionField,
				"store_id": storeID,
			}).Warn("Store not found for order creation")
			return nil, &errors.RecordNotFoundError{Message: errors.StoreNotFound}
		}

		// Foreign key violation
		if pqErr.Code == PqErrCodeForeignKeyViolation {
			logs.WithFields(map[string]interface{}{
				"file":       OrderRepositoryField,
				"function":   OrderCreateFunctionField,
				"pg_code":    pqErr.Code,
				"pg_message": pqErr.Message,
			}).Warn("Foreign key violation during order creation")
			return nil, &errors.ValidationError{Message: pqErr.Message}
		}

		logs.WithFields(map[string]interface{}{
			"file":       OrderRepositoryField,
			"function":   OrderCreateFunctionField,
			"pg_code":    pqErr.Code,
			"pg_message": pqErr.Message,
			"pg_detail":  pqErr.Detail,
		}).Debug("PostgreSQL error details from stored procedure")

		return nil, fmt.Errorf("stored procedure error: %s", pqErr.Message)
	}

	return nil, fmt.Errorf("database operation failed: %w", err)
}

// GetAllByShopIDWithFilters returns orders for a shop with filters applied.
// Lightweight query - NO items included (for real-time dashboard performance).
// Items count is fetched via efficient subquery.
// Returns limit+1 items for pagination (LIMIT+1 strategy).
//
//nolint:gocyclo // Dynamic query building requires multiple conditional branches
func (r *OrderSQLRepository) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) ([]*models.Order, error) {
	startTime := time.Now()

	// Lightweight query - NO items included (for dashboard performance)
	// Items come from separate GetByID call for full order details
	baseQuery := `
		SELECT
			o.id, o.order_number, o.status,
			o.customer_name, o.customer_phone, o.customer_email, o.customer_address_name,
			o.payment_method_id, o.payment_method_code, o.payment_method_name,
			o.delivery_method_id, o.delivery_method_code, o.delivery_method_name,
			o.subtotal, o.shipping_cost, o.total,
			o.created_at,
			(SELECT COUNT(*) FROM order_items oi WHERE oi.order_id = o.id) AS items_count
		FROM orders o
		WHERE o.store_id = $1`

	// Build dynamic WHERE conditions
	conditions := []string{}
	args := []interface{}{shopID}
	argPos := 2

	// Status filter
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("o.status = $%d", argPos))
		args = append(args, *filters.Status)
		argPos++
	}

	// Date range filters
	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("o.created_at >= $%d", argPos))
		args = append(args, *filters.DateFrom)
		argPos++
	}

	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("o.created_at <= $%d", argPos))
		args = append(args, *filters.DateTo)
		argPos++
	}

	// Search: ILIKE on order_number and customer_name
	// Uses trigram indexes (pg_trgm) for fast partial matching
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := *filters.Search
		conditions = append(conditions, fmt.Sprintf(
			"(o.order_number ILIKE $%d OR o.customer_name ILIKE $%d)",
			argPos, argPos,
		))
		args = append(args, "%"+searchTerm+"%")
		argPos++
	}

	// Keyset pagination: use LastID and LastSortValue for cursor-based pagination
	// These values come from the decoded cursor (decoded in contract layer)
	if filters.LastID != nil {
		sortField := fmt.Sprintf("o.%s", filters.SortBy)

		if filters.SortOrder == models.SortOrderDesc {
			if filters.LastSortValue != nil {
				conditions = append(conditions, fmt.Sprintf(
					"(%s < $%d OR (%s = $%d AND o.id < $%d))",
					sortField, argPos, sortField, argPos, argPos+1,
				))
				args = append(args, filters.LastSortValue, *filters.LastID)
				argPos += 2
			} else {
				conditions = append(conditions, fmt.Sprintf("o.id < $%d", argPos))
				args = append(args, *filters.LastID)
				argPos++
			}
		} else {
			if filters.LastSortValue != nil {
				conditions = append(conditions, fmt.Sprintf(
					"(%s > $%d OR (%s = $%d AND o.id > $%d))",
					sortField, argPos, sortField, argPos, argPos+1,
				))
				args = append(args, filters.LastSortValue, *filters.LastID)
				argPos += 2
			} else {
				conditions = append(conditions, fmt.Sprintf("o.id > $%d", argPos))
				args = append(args, *filters.LastID)
				argPos++
			}
		}
	}

	// Append all WHERE conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// ORDER BY with validated and sanitized fields
	// Always include id as secondary sort for stable pagination
	baseQuery += fmt.Sprintf(" ORDER BY o.%s %s, o.id %s",
		filters.SortBy,
		strings.ToUpper(filters.SortOrder),
		strings.ToUpper(filters.SortOrder))

	// LIMIT - request limit + 1 to detect if there are more pages
	baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, filters.Limit+1)

	// Execute query
	logs.WithFields(map[string]interface{}{
		"file":       OrderRepositoryField,
		"function":   OrderGetAllByShopIDWithFiltersFuncField,
		"shop_id":    shopID,
		"sort_by":    filters.SortBy,
		"sort_order": filters.SortOrder,
		"has_search": filters.Search != nil,
		"has_status": filters.Status != nil,
		"has_date":   filters.DateFrom != nil || filters.DateTo != nil,
		"limit":      filters.Limit,
		"last_id":    filters.LastID,
	}).Debug("Executing lightweight order search query (no items)")

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        OrderRepositoryField,
			"function":    OrderGetAllByShopIDWithFiltersFuncField,
			"duration_ms": time.Since(startTime).Milliseconds(),
			"error":       err.Error(),
		}).Error("Failed to query orders with filters")
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	var orders []*models.Order

	for rows.Next() {
		order := &models.Order{
			Customer:       &models.Customer{},
			PaymentMethod:  &models.PaymentMethod{},
			DeliveryMethod: &models.DeliveryMethod{},
		}

		var customerName string
		var customerPhone, customerEmail, customerAddressName sql.NullString
		var paymentMethodID sql.NullInt64
		var paymentMethodCode, paymentMethodName sql.NullString
		var deliveryMethodID sql.NullInt64
		var deliveryMethodCode, deliveryMethodName sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.Status,
			&customerName,
			&customerPhone,
			&customerEmail,
			&customerAddressName,
			&paymentMethodID,
			&paymentMethodCode,
			&paymentMethodName,
			&deliveryMethodID,
			&deliveryMethodCode,
			&deliveryMethodName,
			&order.Subtotal,
			&order.ShippingCost,
			&order.Total,
			&order.CreatedAt,
			&order.ItemsCount,
		)
		if err != nil {
			logs.WithFields(map[string]interface{}{
				"file":     OrderRepositoryField,
				"function": OrderGetAllByShopIDWithFiltersFuncField,
				"error":    err.Error(),
			}).Error("Failed to scan order row")
			return nil, fmt.Errorf("database operation failed")
		}

		order.Customer.Name = customerName
		if customerPhone.Valid {
			order.Customer.Phone = customerPhone.String
		}
		if customerEmail.Valid {
			order.Customer.Email = customerEmail.String
		}
		if customerAddressName.Valid {
			order.Customer.Address = &models.Address{Name: customerAddressName.String}
		}
		if paymentMethodName.Valid {
			order.PaymentMethod.ID = int(paymentMethodID.Int64)
			order.PaymentMethod.Code = models.PaymentMethodCode(paymentMethodCode.String)
			order.PaymentMethod.Name = paymentMethodName.String
		} else {
			order.PaymentMethod = nil
		}
		if deliveryMethodName.Valid {
			order.DeliveryMethod.ID = int(deliveryMethodID.Int64)
			order.DeliveryMethod.Code = models.DeliveryMethodCode(deliveryMethodCode.String)
			order.DeliveryMethod.Name = deliveryMethodName.String
		} else {
			order.DeliveryMethod = nil
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderGetAllByShopIDWithFiltersFuncField,
			"error":    err.Error(),
		}).Error("Error iterating order rows")
		return nil, fmt.Errorf("database operation failed")
	}

	queryDuration := time.Since(startTime).Milliseconds()

	if queryDuration > 50 {
		logs.WithFields(map[string]interface{}{
			"file":         OrderRepositoryField,
			"function":     OrderGetAllByShopIDWithFiltersFuncField,
			"duration_ms":  queryDuration,
			"result_count": len(orders),
			"shop_id":      shopID,
		}).Warn("Slow order query detected")
	}

	logs.WithFields(map[string]interface{}{
		"file":         OrderRepositoryField,
		"function":     OrderGetAllByShopIDWithFiltersFuncField,
		"duration_ms":  queryDuration,
		"result_count": len(orders),
		"shop_id":      shopID,
	}).Debug("Orders query with filters completed")

	return orders, nil
}

// CountByShopIDWithFilters returns total count of orders matching filters.
// Same filter conditions as GetAllByShopIDWithFilters, without cursor/ORDER BY/LIMIT.
func (r *OrderSQLRepository) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) (int, error) {
	startTime := time.Now()

	countQuery := "SELECT COUNT(*) FROM orders o WHERE o.store_id = $1"
	args := []interface{}{shopID}
	argPos := 2

	var conditions []string

	// Status filter
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("o.status = $%d", argPos))
		args = append(args, *filters.Status)
		argPos++
	}

	// Date range filters
	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("o.created_at >= $%d", argPos))
		args = append(args, *filters.DateFrom)
		argPos++
	}

	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("o.created_at <= $%d", argPos))
		args = append(args, *filters.DateTo)
		argPos++
	}

	// Search filter
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := *filters.Search
		conditions = append(conditions, fmt.Sprintf(
			"(o.order_number ILIKE $%d OR o.customer_name ILIKE $%d)",
			argPos, argPos,
		))
		args = append(args, "%"+searchTerm+"%")
	}

	// Append conditions
	if len(conditions) > 0 {
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&count)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderCountByShopIDWithFiltersFuncField,
			"error":    err.Error(),
		}).Error("Failed to count orders")
		return 0, fmt.Errorf("database operation failed")
	}

	queryDuration := time.Since(startTime).Milliseconds()

	logs.WithFields(map[string]interface{}{
		"file":        OrderRepositoryField,
		"function":    OrderCountByShopIDWithFiltersFuncField,
		"duration_ms": queryDuration,
		"count":       count,
		"shop_id":     shopID,
	}).Debug("Orders count completed")

	return count, nil
}
