package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Metrics repository log field constants
const (
	MetricsRepositoryField           = "metrics_repository"
	MetricsRevenueSummaryFuncField   = "get_revenue_summary"
	MetricsOrderStatusDistFuncField  = "get_order_status_distribution"
	MetricsAvgCompletionFuncField    = "get_avg_completion_time"
	MetricsTopProductsFuncField      = "get_top_products"
	MetricsCustomerOverviewFuncField = "get_customer_overview"
	MetricsTopCustomersFuncField     = "get_top_customers"
	MetricsPaymentDistFuncField      = "get_payment_method_distribution"
	MetricsDeliveryDistFuncField     = "get_delivery_method_distribution"
	MetricsRevenueTrendFuncField     = "get_revenue_trend"
	MetricsShippingSummaryFuncField  = "get_shipping_summary"
)

type MetricsSQLRepository struct {
	db *sql.DB
}

func NewMetricsRepository(dataBaseConnection DataBaseConnection) ports.MetricsRepository {
	return &MetricsSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

// GetRevenueSummary returns revenue aggregation for completed orders in a date range.
func (r *MetricsSQLRepository) GetRevenueSummary(ctx context.Context, shopID int, from, to time.Time) (models.RevenueSummary, error) {
	query := `
		SELECT COALESCE(SUM(subtotal - discount), 0), COUNT(*)
		FROM orders
		WHERE store_id = $1
		  AND status = 'completed'
		  AND created_at >= $2
		  AND created_at < $3`

	var summary models.RevenueSummary
	err := r.db.QueryRowContext(ctx, query, shopID, from, to).Scan(&summary.TotalRevenue, &summary.OrderCount)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsRevenueSummaryFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying revenue summary")
		return models.RevenueSummary{}, fmt.Errorf("error querying revenue summary: %w", err)
	}

	if summary.OrderCount > 0 {
		summary.AOV = summary.TotalRevenue / float64(summary.OrderCount)
	}

	return summary, nil
}

// GetRevenueSummaryBatch returns revenue summaries for multiple periods in a single query
// using PostgreSQL FILTER clauses for efficient single-pass aggregation.
func (r *MetricsSQLRepository) GetRevenueSummaryBatch(ctx context.Context, shopID int, periods []models.RevenuePeriod) ([]models.RevenueSummary, error) {
	if len(periods) == 0 {
		return []models.RevenueSummary{}, nil
	}

	// Build query dynamically: 2 columns (revenue, count) per period
	// Parameters: $1=shopID, then $2/$3 for period 0, $4/$5 for period 1, etc.
	// Last two params are the global min(from) and max(to) for index scan bounds.
	var selectCols string
	args := []interface{}{shopID}
	paramIdx := 2

	minFrom := periods[0].From
	maxTo := periods[0].To

	for i, p := range periods {
		if i > 0 {
			selectCols += ",\n\t\t       "
		}
		selectCols += fmt.Sprintf(
			"COALESCE(SUM(subtotal - discount) FILTER (WHERE created_at >= $%d AND created_at < $%d), 0),\n\t\t       COUNT(*) FILTER (WHERE created_at >= $%d AND created_at < $%d)",
			paramIdx, paramIdx+1, paramIdx, paramIdx+1,
		)
		args = append(args, p.From, p.To)
		paramIdx += 2

		if p.From.Before(minFrom) {
			minFrom = p.From
		}
		if p.To.After(maxTo) {
			maxTo = p.To
		}
	}

	// Global bounds for efficient index scan
	boundsFromIdx := paramIdx
	boundsToIdx := paramIdx + 1
	args = append(args, minFrom, maxTo)

	//nolint:gosec // G201: only interpolates column expressions and parameter indices ($N), not user input
	query := fmt.Sprintf(`
		SELECT %s
		FROM orders
		WHERE store_id = $1
		  AND status = 'completed'
		  AND created_at >= $%d
		  AND created_at < $%d`, selectCols, boundsFromIdx, boundsToIdx)

	row := r.db.QueryRowContext(ctx, query, args...)

	// Scan all columns: 2 per period (revenue, count)
	scanDest := make([]interface{}, len(periods)*2)
	summaries := make([]models.RevenueSummary, len(periods))
	for i := range periods {
		scanDest[i*2] = &summaries[i].TotalRevenue
		scanDest[i*2+1] = &summaries[i].OrderCount
	}

	if err := row.Scan(scanDest...); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": "get_revenue_summary_batch",
			"shop_id":  shopID,
			"periods":  len(periods),
			"error":    err.Error(),
		}).Error("Error querying revenue summary batch")
		return nil, fmt.Errorf("error querying revenue summary batch: %w", err)
	}

	// Calculate AOV for each period
	for i := range summaries {
		if summaries[i].OrderCount > 0 {
			summaries[i].AOV = summaries[i].TotalRevenue / float64(summaries[i].OrderCount)
		}
	}

	return summaries, nil
}

// GetOrderStatusDistribution returns order counts grouped by status within a date range.
func (r *MetricsSQLRepository) GetOrderStatusDistribution(ctx context.Context, shopID int, from, to time.Time) ([]models.OrderStatusDistribution, error) {
	query := `
		SELECT status, COUNT(*)
		FROM orders
		WHERE store_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY status
		ORDER BY COUNT(*) DESC`

	rows, err := r.db.QueryContext(ctx, query, shopID, from, to)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsOrderStatusDistFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying order status distribution")
		return nil, fmt.Errorf("error querying order status distribution: %w", err)
	}
	defer rows.Close()

	var distribution []models.OrderStatusDistribution
	for rows.Next() {
		var d models.OrderStatusDistribution
		if err := rows.Scan(&d.Status, &d.Count); err != nil {
			return nil, fmt.Errorf("error scanning order status distribution: %w", err)
		}
		distribution = append(distribution, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order status distribution: %w", err)
	}

	return distribution, nil
}

// GetAvgCompletionTime returns the average hours to complete an order within a date range.
func (r *MetricsSQLRepository) GetAvgCompletionTime(ctx context.Context, shopID int, from, to time.Time) (*float64, error) {
	query := `
		SELECT AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 3600)
		FROM orders
		WHERE store_id = $1
		  AND status = 'completed'
		  AND created_at >= $2
		  AND created_at < $3`

	var avgHours *float64
	err := r.db.QueryRowContext(ctx, query, shopID, from, to).Scan(&avgHours)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsAvgCompletionFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying avg completion time")
		return nil, fmt.Errorf("error querying avg completion time: %w", err)
	}

	return avgHours, nil
}

// GetTopProducts returns top products ranked by quantity or revenue within a date range.
func (r *MetricsSQLRepository) GetTopProducts(ctx context.Context, shopID int, from, to time.Time, sortBy string, limit int) ([]models.TopProduct, error) {
	const topProductsByQuantity = `
		SELECT oi.product_id,
		       (ARRAY_AGG(oi.product_name ORDER BY o.created_at DESC))[1] AS product_name,
		       (ARRAY_AGG(oi.product_image_url ORDER BY o.created_at DESC) FILTER (WHERE oi.product_image_url IS NOT NULL))[1] AS product_image_url,
		       SUM(oi.quantity) AS quantity_sold,
		       SUM(oi.total_price * (o.subtotal - o.discount) / NULLIF(o.subtotal, 0)) AS revenue
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.store_id = $1
		  AND o.status = 'completed'
		  AND o.created_at >= $2
		  AND o.created_at < $3
		  AND oi.product_id IS NOT NULL
		GROUP BY oi.product_id
		ORDER BY SUM(oi.quantity) DESC
		LIMIT $4`

	const topProductsByRevenue = `
		SELECT oi.product_id,
		       (ARRAY_AGG(oi.product_name ORDER BY o.created_at DESC))[1] AS product_name,
		       (ARRAY_AGG(oi.product_image_url ORDER BY o.created_at DESC) FILTER (WHERE oi.product_image_url IS NOT NULL))[1] AS product_image_url,
		       SUM(oi.quantity) AS quantity_sold,
		       SUM(oi.total_price * (o.subtotal - o.discount) / NULLIF(o.subtotal, 0)) AS revenue
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.store_id = $1
		  AND o.status = 'completed'
		  AND o.created_at >= $2
		  AND o.created_at < $3
		  AND oi.product_id IS NOT NULL
		GROUP BY oi.product_id
		ORDER BY SUM(oi.total_price * (o.subtotal - o.discount) / NULLIF(o.subtotal, 0)) DESC
		LIMIT $4`

	query := topProductsByQuantity
	if sortBy == "revenue" {
		query = topProductsByRevenue
	}

	rows, err := r.db.QueryContext(ctx, query, shopID, from, to, limit)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsTopProductsFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying top products")
		return nil, fmt.Errorf("error querying top products: %w", err)
	}
	defer rows.Close()

	var products []models.TopProduct
	for rows.Next() {
		var p models.TopProduct
		var imageURL sql.NullString
		if err := rows.Scan(&p.ProductID, &p.ProductName, &imageURL, &p.QuantitySold, &p.Revenue); err != nil {
			return nil, fmt.Errorf("error scanning top product: %w", err)
		}
		if imageURL.Valid {
			p.ImageURL = imageURL.String
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top products: %w", err)
	}

	return products, nil
}

// GetCustomerOverview returns customer aggregation stats within a date range.
func (r *MetricsSQLRepository) GetCustomerOverview(ctx context.Context, shopID int, from, to time.Time) (models.CustomerOverview, error) {
	query := `
		WITH customer_orders AS (
			SELECT customer_phone, COUNT(*) AS order_count
			FROM orders
			WHERE store_id = $1
			  AND created_at >= $2
			  AND created_at < $3
			  AND customer_phone IS NOT NULL
			  AND customer_phone != ''
			GROUP BY customer_phone
		)
		SELECT
			COUNT(*) AS total_unique,
			COUNT(*) FILTER (WHERE order_count = 1) AS new_count,
			COUNT(*) FILTER (WHERE order_count > 1) AS returning
		FROM customer_orders`

	var overview models.CustomerOverview
	err := r.db.QueryRowContext(ctx, query, shopID, from, to).Scan(
		&overview.TotalUnique, &overview.NewCount, &overview.Returning,
	)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsCustomerOverviewFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying customer overview")
		return models.CustomerOverview{}, fmt.Errorf("error querying customer overview: %w", err)
	}

	return overview, nil
}

// GetTopCustomers returns top customers ranked by total spend within a date range.
func (r *MetricsSQLRepository) GetTopCustomers(ctx context.Context, shopID int, from, to time.Time, limit int) ([]models.TopCustomer, error) {
	query := `
		SELECT
			(ARRAY_AGG(customer_name ORDER BY created_at DESC))[1] AS customer_name,
			customer_phone,
			COUNT(*) AS order_count,
			SUM(subtotal - discount) AS total_spent
		FROM orders
		WHERE store_id = $1
		  AND status = 'completed'
		  AND created_at >= $2
		  AND created_at < $3
		  AND customer_phone IS NOT NULL
		  AND customer_phone != ''
		GROUP BY customer_phone
		ORDER BY total_spent DESC
		LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, shopID, from, to, limit)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsTopCustomersFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying top customers")
		return nil, fmt.Errorf("error querying top customers: %w", err)
	}
	defer rows.Close()

	var customers []models.TopCustomer
	for rows.Next() {
		var c models.TopCustomer
		if err := rows.Scan(&c.CustomerName, &c.CustomerPhone, &c.OrderCount, &c.TotalSpent); err != nil {
			return nil, fmt.Errorf("error scanning top customer: %w", err)
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top customers: %w", err)
	}

	return customers, nil
}

// GetPaymentMethodDistribution returns order counts grouped by payment method within a date range.
func (r *MetricsSQLRepository) GetPaymentMethodDistribution(ctx context.Context, shopID int, from, to time.Time) ([]models.MethodDistribution, error) {
	query := `
		SELECT payment_method_name, payment_method_code, COUNT(*)
		FROM orders
		WHERE store_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		  AND payment_method_code IS NOT NULL
		GROUP BY payment_method_name, payment_method_code
		ORDER BY COUNT(*) DESC`

	rows, err := r.db.QueryContext(ctx, query, shopID, from, to)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsPaymentDistFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying payment method distribution")
		return nil, fmt.Errorf("error querying payment method distribution: %w", err)
	}
	defer rows.Close()

	var distribution []models.MethodDistribution
	for rows.Next() {
		var d models.MethodDistribution
		if err := rows.Scan(&d.Name, &d.Code, &d.OrderCount); err != nil {
			return nil, fmt.Errorf("error scanning payment method distribution: %w", err)
		}
		distribution = append(distribution, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment method distribution: %w", err)
	}

	return distribution, nil
}

// GetDeliveryMethodDistribution returns order counts grouped by delivery method within a date range.
func (r *MetricsSQLRepository) GetDeliveryMethodDistribution(ctx context.Context, shopID int, from, to time.Time) ([]models.MethodDistribution, error) {
	query := `
		SELECT delivery_method_name, delivery_method_code, COUNT(*)
		FROM orders
		WHERE store_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		  AND delivery_method_code IS NOT NULL
		GROUP BY delivery_method_name, delivery_method_code
		ORDER BY COUNT(*) DESC`

	rows, err := r.db.QueryContext(ctx, query, shopID, from, to)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsDeliveryDistFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying delivery method distribution")
		return nil, fmt.Errorf("error querying delivery method distribution: %w", err)
	}
	defer rows.Close()

	var distribution []models.MethodDistribution
	for rows.Next() {
		var d models.MethodDistribution
		if err := rows.Scan(&d.Name, &d.Code, &d.OrderCount); err != nil {
			return nil, fmt.Errorf("error scanning delivery method distribution: %w", err)
		}
		distribution = append(distribution, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating delivery method distribution: %w", err)
	}

	return distribution, nil
}

// GetRevenueTrend returns daily revenue data points within a date range.
func (r *MetricsSQLRepository) GetRevenueTrend(ctx context.Context, shopID int, from, to time.Time, tz string) ([]models.RevenueTrendPoint, error) {
	query := `
		SELECT
			date_trunc('day', created_at AT TIME ZONE $4)::date AS day,
			COALESCE(SUM(subtotal - discount), 0) AS revenue,
			COUNT(*) AS order_count
		FROM orders
		WHERE store_id = $1
		  AND status = 'completed'
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY day
		ORDER BY day ASC`

	rows, err := r.db.QueryContext(ctx, query, shopID, from, to, tz)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsRevenueTrendFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying revenue trend")
		return nil, fmt.Errorf("error querying revenue trend: %w", err)
	}
	defer rows.Close()

	var trend []models.RevenueTrendPoint
	for rows.Next() {
		var p models.RevenueTrendPoint
		var day time.Time
		if err := rows.Scan(&day, &p.Revenue, &p.OrderCount); err != nil {
			return nil, fmt.Errorf("error scanning revenue trend point: %w", err)
		}
		p.Date = day.Format("2006-01-02")
		trend = append(trend, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating revenue trend: %w", err)
	}

	return trend, nil
}

// GetShippingSummary returns aggregated shipping metrics for completed orders with shipping in a date range.
func (r *MetricsSQLRepository) GetShippingSummary(ctx context.Context, shopID int, from, to time.Time) (models.ShippingSummary, error) {
	query := `
		SELECT COALESCE(SUM(shipping_cost), 0), COUNT(*)
		FROM orders
		WHERE store_id = $1
		  AND status = 'completed'
		  AND shipping_cost > 0
		  AND created_at >= $2
		  AND created_at < $3`

	var summary models.ShippingSummary
	err := r.db.QueryRowContext(ctx, query, shopID, from, to).Scan(&summary.TotalShippingRevenue, &summary.OrderCount)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsRepositoryField,
			"function": MetricsShippingSummaryFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error querying shipping summary")
		return models.ShippingSummary{}, fmt.Errorf("error querying shipping summary: %w", err)
	}

	if summary.OrderCount > 0 {
		summary.AverageShippingCost = summary.TotalShippingRevenue / float64(summary.OrderCount)
	}

	return summary, nil
}
