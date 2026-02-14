package order

import (
	"context"
	"sync"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// GetAllOrdersByShopIDUseCase orchestrates the order retrieval with filters and pagination.
// This is the orchestrator layer - coordinates multiple services (NOT repositories directly).
type GetAllOrdersByShopIDUseCase struct {
	shopService       ports.ShopService
	orderService      ports.OrderService
	paginationService ports.PaginationService[*models.Order]
}

func NewGetAllOrdersByShopIDUseCase(
	shopService ports.ShopService,
	orderService ports.OrderService,
	paginationService ports.PaginationService[*models.Order],
) ports.GetAllOrdersByShopIDUseCase {
	return &GetAllOrdersByShopIDUseCase{
		shopService:       shopService,
		orderService:      orderService,
		paginationService: paginationService,
	}
}

// Execute orchestrates the complete flow:
//  1. Validates and normalizes filters (immutable - returns a copy)
//  2. Resolves date filters to shop's timezone (if dates present)
//  3. Fetches total count on first page only (delegates to OrderService)
//  4. Fetches orders with filters (delegates to OrderService)
//  5. Applies pagination logic (delegates to PaginationService)
//
// ShopID is a context parameter (not a filter), passed separately.
//
// Returns:
//   - orders: List of orders (max of 'limit' items)
//   - nextCursor: Opaque cursor for next page (empty if no more pages)
//   - hasMore: true if there are more pages
//   - totalCount: Total count of items (only on first page when LastID is nil, otherwise nil)
//   - error: Any error that occurred
func (uc *GetAllOrdersByShopIDUseCase) Execute(
	ctx context.Context,
	shopID int,
	filters models.OrderFilters,
) ([]*models.Order, string, bool, *int, error) {
	// 1. Validate and normalize filters (immutable - returns a validated copy)
	validatedFilters, err := filters.Validated()
	if err != nil {
		return nil, "", false, nil, err
	}

	// 2. Resolve date filters to shop's timezone
	validatedFilters = uc.resolveDateTimezone(ctx, shopID, validatedFilters)

	var (
		totalCount *int
		orders     []*models.Order
		fetchErr   error
	)

	// 3. First page: execute COUNT and SELECT in parallel to reduce latency
	if validatedFilters.LastID == nil {
		var (
			wg           sync.WaitGroup
			count        int
			countErr     error
			ordersResult []*models.Order
			ordersErr    error
		)

		// Query 1: COUNT (parallel)
		wg.Go(func() {
			count, countErr = uc.orderService.CountByShopIDWithFilters(ctx, shopID, validatedFilters)
		})

		// Query 2: SELECT orders (parallel)
		wg.Go(func() {
			ordersResult, ordersErr = uc.orderService.GetAllByShopIDWithFilters(ctx, shopID, validatedFilters)
		})

		wg.Wait()

		// Assign results after both goroutines complete (no race)
		orders = ordersResult
		fetchErr = ordersErr
		if countErr == nil {
			totalCount = &count
		}
	} else {
		// Subsequent pages: only fetch orders (no count needed)
		orders, fetchErr = uc.orderService.GetAllByShopIDWithFilters(ctx, shopID, validatedFilters)
	}

	if fetchErr != nil {
		return nil, "", false, nil, fetchErr
	}

	// 4. Apply pagination logic (delegated to PaginationService)
	trimmedOrders, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		orders,
		validatedFilters.Limit,
		validatedFilters.SortBy,
	)

	return trimmedOrders, nextCursor, hasMore, totalCount, nil
}

// resolveDateTimezone re-interprets date filters in the shop's timezone
// so that "Feb 10" means the full day in the shop's local time, not UTC.
//
// Priority:
//  1. Uses filters.Timezone (sent by frontend) to avoid a DB roundtrip
//  2. Falls back to loading the shop's timezone from DB
//  3. Falls back to UTC if both fail
func (uc *GetAllOrdersByShopIDUseCase) resolveDateTimezone(
	ctx context.Context,
	shopID int,
	filters models.OrderFilters,
) models.OrderFilters {
	if filters.DateFrom == nil && filters.DateTo == nil {
		return filters
	}

	// 1. Try timezone from frontend (avoids DB roundtrip)
	if filters.Timezone != nil && *filters.Timezone != "" {
		if loc, err := time.LoadLocation(*filters.Timezone); err == nil {
			return filters.WithTimezone(loc)
		}
		logs.WithFields(map[string]interface{}{
			"file":     "get_all_orders_by_shop_id",
			"function": "resolve_date_timezone",
			"shop_id":  shopID,
			"timezone": *filters.Timezone,
		}).Warn("Invalid timezone from request, falling back to shop timezone")
	}

	// 2. Fallback: load timezone from shop
	shop, err := uc.shopService.GetByID(ctx, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     "get_all_orders_by_shop_id",
			"function": "resolve_date_timezone",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Warn("Failed to load shop timezone, using UTC for date filters")
		return filters
	}

	if shop.Timezone == nil || shop.Timezone.Identifier == "" {
		return filters
	}

	loc, err := time.LoadLocation(shop.Timezone.Identifier)
	if err != nil {
		return filters
	}

	return filters.WithTimezone(loc)
}
