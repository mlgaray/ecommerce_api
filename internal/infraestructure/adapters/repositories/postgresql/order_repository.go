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
	OrderGetByIDFuncField                   = "get_by_id"
	OrderUpdateStatusFuncField              = "update_status"
	OrderUpdateFuncField                    = "update"
)

type OrderSQLRepository struct {
	db *sql.DB
}

func NewOrderRepository(dataBaseConnection DataBaseConnection) ports.OrderRepository {
	return &OrderSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

// orderFieldParams holds nullable field values extracted from an order for stored procedure calls.
type orderFieldParams struct {
	customerName, customerPhone, customerEmail *string
	addressName, addressPlaceID                *string
	addressLat, addressLng                     *float64
	paymentMethodID                            *int
	paymentMethodCode, paymentMethodName       *string
	deliveryMethodID                           *int
	deliveryMethodCode, deliveryMethodName     *string
	deliveryZoneID                             *int
	deliveryZoneName                           *string
	deliveryZonePrice                          *float64
	discount                                   float64
	couponCode                                 *string
	couponType                                 *string
	couponValue                                *float64
	couponDiscAmt                              *float64
	couponMinOrder                             *float64
	couponID                                   *int
	couponPhone                                *string
}

// extractOrderFieldParams extracts nullable fields from an order model for stored procedure parameters.
// Used by both Create and Update to avoid code duplication.
func (r *OrderSQLRepository) extractOrderFieldParams(order *models.Order) orderFieldParams {
	var p orderFieldParams

	if order.Customer != nil {
		p.customerName = &order.Customer.Name
		p.customerPhone = &order.Customer.Phone
		p.customerEmail = &order.Customer.Email

		if order.Customer.Address != nil {
			p.addressName = &order.Customer.Address.Name
			p.addressPlaceID = &order.Customer.Address.PlaceID
			p.addressLat = &order.Customer.Address.Lat
			p.addressLng = &order.Customer.Address.Lng
		}
	}

	if order.PaymentMethod != nil {
		p.paymentMethodID = &order.PaymentMethod.ID
		code := string(order.PaymentMethod.Code)
		p.paymentMethodCode = &code
		p.paymentMethodName = &order.PaymentMethod.Name
	}

	if order.DeliveryMethod != nil {
		p.deliveryMethodID = &order.DeliveryMethod.ID
		code := string(order.DeliveryMethod.Code)
		p.deliveryMethodCode = &code
		p.deliveryMethodName = &order.DeliveryMethod.Name

		if len(order.DeliveryMethod.DeliveryZones) > 0 {
			zone := order.DeliveryMethod.DeliveryZones[0]
			p.deliveryZoneID = &zone.ID
			p.deliveryZoneName = &zone.Name
			p.deliveryZonePrice = &zone.Price
		}
	}

	p.discount = order.Discount

	if order.Coupon != nil {
		p.couponCode = &order.Coupon.Code
		couponType := string(order.Coupon.Type)
		p.couponType = &couponType
		p.couponValue = &order.Coupon.Value
		p.couponDiscAmt = &order.Coupon.DiscountAmount
		p.couponMinOrder = &order.Coupon.MinOrderAmount
		p.couponID = &order.Coupon.ID
		if order.Customer != nil && order.Customer.Phone != "" {
			p.couponPhone = &order.Customer.Phone
		}
	}

	return p
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

	// 2. Extract nullable fields from order
	p := r.extractOrderFieldParams(order)

	// 3. Call stored procedure
	var resultJSON []byte
	err = r.db.QueryRowContext(ctx, `
		SELECT create_order(
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29
		)`,
		order.Store.ID,
		order.Store.Name,
		order.Store.Slug,
		p.customerName,
		p.customerPhone,
		p.customerEmail,
		p.addressName,
		p.addressPlaceID,
		p.addressLat,
		p.addressLng,
		p.paymentMethodID,
		p.paymentMethodCode,
		p.paymentMethodName,
		p.deliveryMethodID,
		p.deliveryMethodCode,
		p.deliveryMethodName,
		p.deliveryZoneID,
		p.deliveryZoneName,
		p.deliveryZonePrice,
		order.ShippingCost,
		p.discount,
		p.couponCode,
		p.couponType,
		p.couponValue,
		p.couponDiscAmt,
		p.couponMinOrder,
		p.couponID,
		p.couponPhone,
		itemsJSON,
	).Scan(&resultJSON)

	if err != nil {
		return r.handleCreateError(err, order.Store.ID)
	}

	// 8. Parse result
	var result struct {
		ID           int       `json:"id"`
		OrderNumber  string    `json:"order_number"`
		Status       string    `json:"status"`
		Subtotal     float64   `json:"subtotal"`
		Discount     float64   `json:"discount"`
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

	// 9. Update order with DB-generated values
	order.ID = result.ID
	order.OrderNumber = result.OrderNumber
	order.Status = models.OrderStatus(result.Status)
	order.Subtotal = result.Subtotal
	order.Discount = result.Discount
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
		Quantity    int     `json:"quantity"`
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
		Notes            string               `json:"notes,omitempty"`
		SelectedOptions  []selectedOptionJSON `json:"selected_options,omitempty"`
	}

	jsonItems := make([]itemJSON, 0, len(items))
	for _, item := range items {
		jItem := itemJSON{
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
			Notes:      item.Notes,
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
					quantity := option.Quantity
					if quantity <= 0 {
						quantity = 1
					}
					jItem.SelectedOptions = append(jItem.SelectedOptions, selectedOptionJSON{
						VariantID:   variant.ID,
						OptionID:    option.ID,
						VariantName: variant.Name,
						OptionName:  option.Name,
						OptionPrice: option.Price,
						Quantity:    quantity,
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

		// Coupon usage limit errors (P0003 from SP)
		if pqErr.Code == PqErrCodeRaiseException {
			return nil, mapCouponLimitError(pqErr.Message)
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

// mapCouponLimitError maps SP coupon limit error messages to domain errors.
func mapCouponLimitError(message string) error {
	switch message {
	case errors.CouponUsageLimitReached:
		return &errors.BusinessRuleError{Message: errors.CouponUsageLimitReached}
	case errors.CouponPhoneLimitReached:
		return &errors.BusinessRuleError{Message: errors.CouponPhoneLimitReached}
	default:
		return &errors.BusinessRuleError{Message: message}
	}
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
			o.delivery_zone_name,
			o.subtotal, o.shipping_cost, o.discount, o.total,
			o.coupon_code, o.coupon_type, o.coupon_value,
			o.coupon_discount_amount, o.coupon_min_order_amount,
			o.created_at, o.updated_at,
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
		var deliveryZoneName sql.NullString
		var couponCode, couponType sql.NullString
		var couponValue, couponDiscAmt, couponMinOrder sql.NullFloat64

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
			&deliveryZoneName,
			&order.Subtotal,
			&order.ShippingCost,
			&order.Discount,
			&order.Total,
			&couponCode,
			&couponType,
			&couponValue,
			&couponDiscAmt,
			&couponMinOrder,
			&order.CreatedAt,
			&order.UpdatedAt,
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

			// Attach delivery zone name if present (lightweight - only name for list views)
			if deliveryZoneName.Valid {
				order.DeliveryMethod.DeliveryZones = []*models.DeliveryZone{
					{Name: deliveryZoneName.String},
				}
			}
		} else {
			order.DeliveryMethod = nil
		}

		// Populate coupon snapshot (lightweight — no usage_count needed in list)
		if couponCode.Valid {
			order.Coupon = &models.Coupon{
				Code:           couponCode.String,
				Type:           models.CouponType(couponType.String),
				Value:          couponValue.Float64,
				DiscountAmount: couponDiscAmt.Float64,
				MinOrderAmount: couponMinOrder.Float64,
			}
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

// GetByID retrieves a single order with full details (items, selected options).
// Uses CTE-based query with JSON aggregation for items and selections.
// Filters by both order_id AND store_id for security.
//
//nolint:gocyclo // Complexity is intentional - linear field mappings for nullable columns (customer, payment, delivery, address)
func (r *OrderSQLRepository) GetByID(ctx context.Context, shopID int, orderID int) (*models.Order, error) {
	startTime := time.Now()

	query := `
		WITH items_with_selections AS (
			SELECT
				oi.id, oi.order_id,
				oi.product_id, oi.product_name, oi.product_image_url,
				oi.base_price, oi.is_promotional, oi.promotional_price,
				oi.quantity, oi.unit_price, oi.total_price, oi.notes,
				COALESCE(
					(SELECT json_agg(json_build_object(
						'variant_id', ois.variant_id,
						'variant_name', ois.variant_name,
						'option_id', ois.option_id,
						'option_name', ois.option_name,
						'option_price', ois.option_price,
						'quantity', ois.quantity
					) ORDER BY ois.id)
					FROM order_item_selections ois
					WHERE ois.order_item_id = oi.id),
					'[]'::json
				) AS selected_options
			FROM order_items oi
			WHERE oi.order_id = $1
		),
		items_agg AS (
			SELECT
				iws.order_id,
				json_agg(json_build_object(
					'id', iws.id,
					'product_id', iws.product_id,
					'product_name', iws.product_name,
					'product_image_url', iws.product_image_url,
					'base_price', iws.base_price,
					'is_promotional', iws.is_promotional,
					'promotional_price', iws.promotional_price,
					'quantity', iws.quantity,
					'unit_price', iws.unit_price,
					'total_price', iws.total_price,
					'notes', iws.notes,
					'selected_options', iws.selected_options
				) ORDER BY iws.id) AS items
			FROM items_with_selections iws
			GROUP BY iws.order_id
		)
		SELECT
			o.id, o.order_number, o.status,
			o.store_id, o.store_name, o.store_slug,
			o.customer_name, o.customer_phone, o.customer_email,
			o.customer_address_name, o.customer_address_place_id,
			o.customer_address_lat, o.customer_address_lng,
			o.payment_method_id, o.payment_method_code, o.payment_method_name,
			o.delivery_method_id, o.delivery_method_code, o.delivery_method_name,
			o.delivery_zone_id, o.delivery_zone_name, o.delivery_zone_price,
			o.subtotal, o.shipping_cost, o.discount, o.total,
			o.coupon_code, o.coupon_type, o.coupon_value,
			o.coupon_discount_amount, o.coupon_min_order_amount,
			o.created_at, o.updated_at,
			COALESCE(ia.items, '[]'::json) AS items
		FROM orders o
		LEFT JOIN items_agg ia ON ia.order_id = o.id
		WHERE o.id = $1 AND o.store_id = $2`

	var (
		customerName                                string
		customerPhone, customerEmail                sql.NullString
		customerAddressName, customerAddressPlaceID sql.NullString
		customerAddressLat, customerAddressLng      sql.NullFloat64
		paymentMethodID                             sql.NullInt64
		paymentMethodCode, paymentMethodName        sql.NullString
		deliveryMethodID                            sql.NullInt64
		deliveryMethodCode, deliveryMethodName      sql.NullString
		deliveryZoneID                              sql.NullInt64
		deliveryZoneName                            sql.NullString
		deliveryZonePrice                           sql.NullFloat64
		couponCode, couponType                      sql.NullString
		couponValue, couponDiscAmt, couponMinOrder  sql.NullFloat64
		storeID                                     int
		storeName, storeSlug                        string
		itemsJSON                                   []byte
	)

	order := &models.Order{}

	err := r.db.QueryRowContext(ctx, query, orderID, shopID).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.Status,
		&storeID,
		&storeName,
		&storeSlug,
		&customerName,
		&customerPhone,
		&customerEmail,
		&customerAddressName,
		&customerAddressPlaceID,
		&customerAddressLat,
		&customerAddressLng,
		&paymentMethodID,
		&paymentMethodCode,
		&paymentMethodName,
		&deliveryMethodID,
		&deliveryMethodCode,
		&deliveryMethodName,
		&deliveryZoneID,
		&deliveryZoneName,
		&deliveryZonePrice,
		&order.Subtotal,
		&order.ShippingCost,
		&order.Discount,
		&order.Total,
		&couponCode,
		&couponType,
		&couponValue,
		&couponDiscAmt,
		&couponMinOrder,
		&order.CreatedAt,
		&order.UpdatedAt,
		&itemsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logs.WithFields(map[string]interface{}{
				"file":     OrderRepositoryField,
				"function": OrderGetByIDFuncField,
				"order_id": orderID,
				"shop_id":  shopID,
			}).Warn("Order not found")
			return nil, &errors.RecordNotFoundError{Message: errors.OrderNotFound}
		}
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderGetByIDFuncField,
			"order_id": orderID,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to query order by ID")
		return nil, fmt.Errorf("database operation failed")
	}

	// Populate store
	order.Store = &models.Store{
		ID:   storeID,
		Name: storeName,
		Slug: storeSlug,
	}

	// Populate customer
	order.Customer = &models.Customer{Name: customerName}
	if customerPhone.Valid {
		order.Customer.Phone = customerPhone.String
	}
	if customerEmail.Valid {
		order.Customer.Email = customerEmail.String
	}
	if customerAddressName.Valid {
		order.Customer.Address = &models.Address{
			Name: customerAddressName.String,
		}
		if customerAddressPlaceID.Valid {
			order.Customer.Address.PlaceID = customerAddressPlaceID.String
		}
		if customerAddressLat.Valid {
			order.Customer.Address.Lat = customerAddressLat.Float64
		}
		if customerAddressLng.Valid {
			order.Customer.Address.Lng = customerAddressLng.Float64
		}
	}

	// Populate payment method
	if paymentMethodName.Valid {
		order.PaymentMethod = &models.PaymentMethod{
			ID:   int(paymentMethodID.Int64),
			Code: models.PaymentMethodCode(paymentMethodCode.String),
			Name: paymentMethodName.String,
		}
	}

	// Populate delivery method
	if deliveryMethodName.Valid {
		order.DeliveryMethod = &models.DeliveryMethod{
			ID:   int(deliveryMethodID.Int64),
			Code: models.DeliveryMethodCode(deliveryMethodCode.String),
			Name: deliveryMethodName.String,
		}
	}

	// Populate delivery zone (attached to delivery method)
	if deliveryZoneName.Valid && order.DeliveryMethod != nil {
		order.DeliveryMethod.DeliveryZones = []*models.DeliveryZone{
			{
				ID:    int(deliveryZoneID.Int64),
				Name:  deliveryZoneName.String,
				Price: deliveryZonePrice.Float64,
			},
		}
	}

	// Populate coupon (snapshot from order)
	if couponCode.Valid {
		order.Coupon = &models.Coupon{
			Code:           couponCode.String,
			Type:           models.CouponType(couponType.String),
			Value:          couponValue.Float64,
			DiscountAmount: couponDiscAmt.Float64,
			MinOrderAmount: couponMinOrder.Float64,
		}
	}

	// Unmarshal items JSON
	items, err := r.unmarshalOrderItems(itemsJSON)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderGetByIDFuncField,
			"sub_func": UnmarshallField,
			"order_id": orderID,
			"error":    err.Error(),
		}).Error("Failed to unmarshal order items")
		return nil, fmt.Errorf("database operation failed")
	}
	order.Items = items
	order.ItemsCount = len(items)

	queryDuration := time.Since(startTime).Milliseconds()

	logs.WithFields(map[string]interface{}{
		"file":        OrderRepositoryField,
		"function":    OrderGetByIDFuncField,
		"order_id":    orderID,
		"shop_id":     shopID,
		"items_count": len(items),
		"duration_ms": queryDuration,
	}).Debug("Order retrieved by ID")

	return order, nil
}

// UpdateStatus updates the order status using optimistic locking.
// The currentStatus in the WHERE clause ensures no concurrent modification occurred.
func (r *OrderSQLRepository) UpdateStatus(ctx context.Context, shopID int, orderID int, currentStatus models.OrderStatus, newStatus models.OrderStatus) error {
	startTime := time.Now()

	query := `
		UPDATE orders SET status = $1, updated_at = now()
		WHERE id = $2 AND store_id = $3 AND status = $4`

	result, err := r.db.ExecContext(ctx, query, newStatus, orderID, shopID, currentStatus)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":           OrderRepositoryField,
			"function":       OrderUpdateStatusFuncField,
			"order_id":       orderID,
			"shop_id":        shopID,
			"current_status": currentStatus,
			"new_status":     newStatus,
			"error":          err.Error(),
		}).Error("Failed to update order status")
		return fmt.Errorf("database operation failed")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderUpdateStatusFuncField,
			"order_id": orderID,
			"error":    err.Error(),
		}).Error("Failed to get rows affected")
		return fmt.Errorf("database operation failed")
	}

	if rowsAffected == 0 {
		logs.WithFields(map[string]interface{}{
			"file":           OrderRepositoryField,
			"function":       OrderUpdateStatusFuncField,
			"order_id":       orderID,
			"shop_id":        shopID,
			"current_status": currentStatus,
			"new_status":     newStatus,
		}).Warn("Concurrent modification detected - status changed between read and write")
		return &errors.ConcurrentModificationError{Message: errors.OrderConcurrentModification}
	}

	logs.WithFields(map[string]interface{}{
		"file":        OrderRepositoryField,
		"function":    OrderUpdateStatusFuncField,
		"order_id":    orderID,
		"shop_id":     shopID,
		"new_status":  newStatus,
		"duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Order status updated successfully")

	return nil
}

// unmarshalOrderItems converts the JSON-aggregated items from the CTE query into domain models.
func (r *OrderSQLRepository) unmarshalOrderItems(itemsJSON []byte) ([]*models.OrderItem, error) {
	type selectedOptionJSON struct {
		VariantID   int     `json:"variant_id"`
		VariantName string  `json:"variant_name"`
		OptionID    int     `json:"option_id"`
		OptionName  string  `json:"option_name"`
		OptionPrice float64 `json:"option_price"`
		Quantity    int     `json:"quantity"`
	}

	type itemJSON struct {
		ID               int                  `json:"id"`
		ProductID        int                  `json:"product_id"`
		ProductName      string               `json:"product_name"`
		ProductImageURL  string               `json:"product_image_url"`
		BasePrice        float64              `json:"base_price"`
		IsPromotional    bool                 `json:"is_promotional"`
		PromotionalPrice float64              `json:"promotional_price"`
		Quantity         int                  `json:"quantity"`
		UnitPrice        float64              `json:"unit_price"`
		TotalPrice       float64              `json:"total_price"`
		Notes            string               `json:"notes"`
		SelectedOptions  []selectedOptionJSON `json:"selected_options"`
	}

	var rawItems []itemJSON
	if err := json.Unmarshal(itemsJSON, &rawItems); err != nil {
		return nil, err
	}

	items := make([]*models.OrderItem, 0, len(rawItems))
	for _, ri := range rawItems {
		item := &models.OrderItem{
			Quantity:   ri.Quantity,
			UnitPrice:  ri.UnitPrice,
			TotalPrice: ri.TotalPrice,
			Notes:      ri.Notes,
			Product: &models.Product{
				ID:               ri.ProductID,
				Name:             ri.ProductName,
				Price:            ri.BasePrice,
				IsPromotional:    ri.IsPromotional,
				PromotionalPrice: ri.PromotionalPrice,
			},
		}

		// Set product image if available
		if ri.ProductImageURL != "" {
			item.Product.Images = []*models.Image{
				{URL: ri.ProductImageURL},
			}
		}

		// Convert selected options to variant/option structure
		for _, so := range ri.SelectedOptions {
			// Find or create the variant
			var variant *models.Variant
			for _, v := range item.Product.Variants {
				if v.ID == so.VariantID {
					variant = v
					break
				}
			}
			if variant == nil {
				variant = &models.Variant{
					ID:   so.VariantID,
					Name: so.VariantName,
				}
				item.Product.Variants = append(item.Product.Variants, variant)
			}
			quantity := so.Quantity
			if quantity <= 0 {
				quantity = 1
			}
			variant.Options = append(variant.Options, &models.Option{
				ID:       so.OptionID,
				Name:     so.OptionName,
				Price:    so.OptionPrice,
				Quantity: quantity,
			})
		}

		items = append(items, item)
	}

	return items, nil
}

// Update updates an order with all its items and selected options.
// Uses stored procedure to handle transactional update.
// Recalculates subtotal and total from items.
// Deletes all existing items (CASCADE handles selections) and inserts new ones.
func (r *OrderSQLRepository) Update(ctx context.Context, shopID int, order *models.Order) error {
	startTime := time.Now()

	// 1. Serialize items to JSONB
	itemsJSON, err := r.serializeItems(order.Items)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderUpdateFuncField,
			"sub_func": "serialize_items",
			"error":    err.Error(),
		}).Error("Failed to serialize order items")
		return fmt.Errorf("database operation failed")
	}

	// 2. Extract nullable fields from order
	p := r.extractOrderFieldParams(order)

	// 3. Call stored procedure
	var resultJSON []byte
	err = r.db.QueryRowContext(ctx, `
		SELECT update_order(
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28
		)`,
		order.ID,
		shopID,
		p.customerName,
		p.customerPhone,
		p.customerEmail,
		p.addressName,
		p.addressPlaceID,
		p.addressLat,
		p.addressLng,
		p.paymentMethodID,
		p.paymentMethodCode,
		p.paymentMethodName,
		p.deliveryMethodID,
		p.deliveryMethodCode,
		p.deliveryMethodName,
		p.deliveryZoneID,
		p.deliveryZoneName,
		p.deliveryZonePrice,
		order.ShippingCost,
		p.discount,
		p.couponCode,
		p.couponType,
		p.couponValue,
		p.couponDiscAmt,
		p.couponMinOrder,
		p.couponID,
		p.couponPhone,
		itemsJSON,
	).Scan(&resultJSON)

	if err != nil {
		return r.handleUpdateError(err, shopID, order.ID)
	}

	// 8. Parse result
	var result struct {
		UpdatedAt time.Time `json:"updated_at"`
		Subtotal  float64   `json:"subtotal"`
		Discount  float64   `json:"discount"`
		Total     float64   `json:"total"`
	}

	if err := json.Unmarshal(resultJSON, &result); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderRepositoryField,
			"function": OrderUpdateFuncField,
			"sub_func": UnmarshallField,
			"error":    err.Error(),
		}).Error("Failed to unmarshal stored procedure result")
		return fmt.Errorf("database operation failed")
	}

	// 9. Update order with DB-calculated values
	order.Subtotal = result.Subtotal
	order.Discount = result.Discount
	order.Total = result.Total
	order.UpdatedAt = result.UpdatedAt

	logs.WithFields(map[string]interface{}{
		"file":              OrderRepositoryField,
		"function":          OrderUpdateFuncField,
		"order_id":          order.ID,
		"store_id":          shopID,
		"items_count":       len(order.Items),
		"total_duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Order update completed (stored procedure)")

	return nil
}

// handleUpdateError processes PostgreSQL errors from the update_order stored procedure.
func (r *OrderSQLRepository) handleUpdateError(err error, storeID int, orderID int) error {
	logs.WithFields(map[string]interface{}{
		"file":     OrderRepositoryField,
		"function": OrderUpdateFuncField,
		"store_id": storeID,
		"order_id": orderID,
		"error":    err.Error(),
	}).Error("Failed to update order via stored procedure")

	// Check if it's a PostgreSQL error
	if pqErr, ok := err.(*pq.Error); ok {
		// Order not found (P0002)
		if pqErr.Code == PqErrCodeNoDataFound {
			logs.WithFields(map[string]interface{}{
				"file":     OrderRepositoryField,
				"function": OrderUpdateFuncField,
				"store_id": storeID,
				"order_id": orderID,
			}).Warn("Order not found for update")
			return &errors.RecordNotFoundError{Message: errors.OrderNotFound}
		}

		// Foreign key violation
		if pqErr.Code == PqErrCodeForeignKeyViolation {
			logs.WithFields(map[string]interface{}{
				"file":       OrderRepositoryField,
				"function":   OrderUpdateFuncField,
				"pg_code":    pqErr.Code,
				"pg_message": pqErr.Message,
			}).Warn("Foreign key violation during order update")
			return &errors.ValidationError{Message: pqErr.Message}
		}

		// Coupon usage limit errors (P0003 from SP)
		if pqErr.Code == PqErrCodeRaiseException {
			return mapCouponLimitError(pqErr.Message)
		}

		logs.WithFields(map[string]interface{}{
			"file":       OrderRepositoryField,
			"function":   OrderUpdateFuncField,
			"pg_code":    pqErr.Code,
			"pg_message": pqErr.Message,
			"pg_detail":  pqErr.Detail,
		}).Debug("PostgreSQL error details from stored procedure")

		return fmt.Errorf("stored procedure error: %s", pqErr.Message)
	}

	return fmt.Errorf("database operation failed: %w", err)
}
