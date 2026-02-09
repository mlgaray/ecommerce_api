package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Order repository log field constants
const (
	OrderRepositoryField     = "order_repository"
	OrderCreateFunctionField = "create"
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
