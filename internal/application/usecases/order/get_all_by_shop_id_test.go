package order

import (
	"context"
	stdErrors "errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// TestMain initializes the global logger required by the use case's logging calls.
func TestMain(m *testing.M) {
	logs.Init()
	os.Exit(m.Run())
}

// =============================================================================
// Test Helpers
// =============================================================================

func newTestOrders() []*models.Order {
	return []*models.Order{
		{ID: 1, OrderNumber: "ORD-001", Total: 100.00, CreatedAt: time.Now()},
		{ID: 2, OrderNumber: "ORD-002", Total: 200.00, CreatedAt: time.Now()},
	}
}

func newTestShopWithTimezone(identifier string) *models.Shop {
	return &models.Shop{
		ID:   1,
		Name: "Test Shop",
		Timezone: &models.Timezone{
			ID:         1,
			Name:       "Argentina",
			Identifier: identifier,
			UTCOffset:  "-03:00",
		},
	}
}

func newTestShopWithoutTimezone() *models.Shop {
	return &models.Shop{
		ID:   1,
		Name: "Test Shop",
	}
}

func strPtr(s string) *string {
	return &s
}

func defaultFilters() models.OrderFilters {
	return models.OrderFilters{}
}

// =============================================================================
// GetAllOrdersByShopIDUseCase Validation Tests
// =============================================================================

func TestGetAllOrdersByShopIDUseCase_Execute_Validation(t *testing.T) {
	shopID := 1

	t.Run("when invalid status then returns ValidationError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.OrderFilters{
			Status: strPtr("invalid_status"),
		}

		shopServiceMock := mocks.NewShopService(t)
		orderServiceMock := mocks.NewOrderService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Order](t)

		useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

		// Act
		orders, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.InvalidOrderStatus, valErr.Message)
		assert.Nil(t, orders)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when invalid sort field then returns ValidationError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.OrderFilters{
			SortBy: "invalid_field",
		}

		shopServiceMock := mocks.NewShopService(t)
		orderServiceMock := mocks.NewOrderService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Order](t)

		useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

		// Act
		orders, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.InvalidSortField, valErr.Message)
		assert.Nil(t, orders)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

// =============================================================================
// GetAllOrdersByShopIDUseCase Timezone Resolution Tests
// =============================================================================

// dateMatcherFor returns a mock.MatchedBy that asserts dates match the expected values.
func dateMatcherFor(expectedFrom, expectedTo time.Time) interface{} {
	return mock.MatchedBy(func(f models.OrderFilters) bool {
		return f.DateFrom != nil && f.DateFrom.Equal(expectedFrom) &&
			f.DateTo != nil && f.DateTo.Equal(expectedTo)
	})
}

// setupOrderServiceWithDateMatcher configures count+get expectations with a date matcher.
func setupOrderServiceWithDateMatcher(
	t *testing.T,
	dateMatcher interface{},
	orders []*models.Order,
) *mocks.OrderService {
	orderServiceMock := mocks.NewOrderService(t)
	orderServiceMock.EXPECT().
		CountByShopIDWithFilters(mock.Anything, mock.Anything, dateMatcher).
		Return(len(orders), nil)
	orderServiceMock.EXPECT().
		GetAllByShopIDWithFilters(mock.Anything, mock.Anything, dateMatcher).
		Return(orders, nil)
	return orderServiceMock
}

func TestGetAllOrdersByShopIDUseCase_Execute_TzFromFilters(t *testing.T) {
	// Arrange
	shopID := 1
	ctx := context.Background()
	dateFrom := time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
	tz := "America/Argentina/Buenos_Aires"
	filters := models.OrderFilters{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Timezone: &tz,
	}

	expectedLoc, _ := time.LoadLocation(tz)
	expectedOrders := newTestOrders()
	matcher := dateMatcherFor(
		time.Date(2025, 2, 10, 0, 0, 0, 0, expectedLoc),
		time.Date(2025, 2, 15, 23, 59, 59, 999999000, expectedLoc),
	)

	shopServiceMock := mocks.NewShopService(t)
	orderServiceMock := setupOrderServiceWithDateMatcher(t, matcher, expectedOrders)

	paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
	paginationServiceMock.EXPECT().
		ApplyPagination(expectedOrders, models.DefaultLimit, models.SortByCreatedAt).
		Return(expectedOrders, "", false)

	useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

	// Act
	orders, _, _, _, err := useCase.Execute(ctx, shopID, filters)

	// Assert - ShopService.GetByID not called (testify fails if unexpected calls)
	require.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
}

func TestGetAllOrdersByShopIDUseCase_Execute_TzFromShop(t *testing.T) {
	// Arrange
	shopID := 1
	ctx := context.Background()
	dateFrom := time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
	filters := models.OrderFilters{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	shopTzIdentifier := "America/Argentina/Buenos_Aires"
	expectedLoc, _ := time.LoadLocation(shopTzIdentifier)
	expectedOrders := newTestOrders()
	matcher := dateMatcherFor(
		time.Date(2025, 2, 10, 0, 0, 0, 0, expectedLoc),
		time.Date(2025, 2, 15, 23, 59, 59, 999999000, expectedLoc),
	)

	shopServiceMock := mocks.NewShopService(t)
	shopServiceMock.EXPECT().
		GetByID(mock.Anything, shopID).
		Return(newTestShopWithTimezone(shopTzIdentifier), nil)

	orderServiceMock := setupOrderServiceWithDateMatcher(t, matcher, expectedOrders)

	paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
	paginationServiceMock.EXPECT().
		ApplyPagination(expectedOrders, models.DefaultLimit, models.SortByCreatedAt).
		Return(expectedOrders, "", false)

	useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

	// Act
	orders, _, _, _, err := useCase.Execute(ctx, shopID, filters)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
}

func TestGetAllOrdersByShopIDUseCase_Execute_InvalidTzFallback(t *testing.T) {
	// Arrange
	shopID := 1
	ctx := context.Background()
	dateFrom := time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
	invalidTz := "Invalid/Timezone"
	filters := models.OrderFilters{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Timezone: &invalidTz,
	}

	shopTzIdentifier := "America/Argentina/Buenos_Aires"
	expectedLoc, _ := time.LoadLocation(shopTzIdentifier)
	expectedOrders := newTestOrders()
	matcher := dateMatcherFor(
		time.Date(2025, 2, 10, 0, 0, 0, 0, expectedLoc),
		time.Date(2025, 2, 15, 23, 59, 59, 999999000, expectedLoc),
	)

	shopServiceMock := mocks.NewShopService(t)
	shopServiceMock.EXPECT().
		GetByID(mock.Anything, shopID).
		Return(newTestShopWithTimezone(shopTzIdentifier), nil)

	orderServiceMock := setupOrderServiceWithDateMatcher(t, matcher, expectedOrders)

	paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
	paginationServiceMock.EXPECT().
		ApplyPagination(expectedOrders, models.DefaultLimit, models.SortByCreatedAt).
		Return(expectedOrders, "", false)

	useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

	// Act
	orders, _, _, _, err := useCase.Execute(ctx, shopID, filters)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
}

func TestGetAllOrdersByShopIDUseCase_Execute_NoShopTz(t *testing.T) {
	// Arrange - shop has no timezone, dates remain in UTC
	shopID := 1
	ctx := context.Background()
	dateFrom := time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
	filters := models.OrderFilters{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	expectedOrders := newTestOrders()
	matcher := dateMatcherFor(dateFrom, dateTo)

	shopServiceMock := mocks.NewShopService(t)
	shopServiceMock.EXPECT().
		GetByID(mock.Anything, shopID).
		Return(newTestShopWithoutTimezone(), nil)

	orderServiceMock := setupOrderServiceWithDateMatcher(t, matcher, expectedOrders)

	paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
	paginationServiceMock.EXPECT().
		ApplyPagination(expectedOrders, models.DefaultLimit, models.SortByCreatedAt).
		Return(expectedOrders, "", false)

	useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

	// Act
	orders, _, _, _, err := useCase.Execute(ctx, shopID, filters)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
}

func TestGetAllOrdersByShopIDUseCase_Execute_NoDates(t *testing.T) {
	// Arrange - no dates = no timezone resolution, no shop fetch
	shopID := 1
	ctx := context.Background()
	filters := defaultFilters()
	expectedOrders := newTestOrders()

	shopServiceMock := mocks.NewShopService(t)

	orderServiceMock := mocks.NewOrderService(t)
	orderServiceMock.EXPECT().
		CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
		Return(2, nil)
	orderServiceMock.EXPECT().
		GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
		Return(expectedOrders, nil)

	paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
	paginationServiceMock.EXPECT().
		ApplyPagination(expectedOrders, models.DefaultLimit, models.SortByCreatedAt).
		Return(expectedOrders, "", false)

	useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

	// Act
	orders, _, _, _, err := useCase.Execute(ctx, shopID, filters)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
}

// =============================================================================
// GetAllOrdersByShopIDUseCase Query Execution Tests
// =============================================================================

func TestGetAllOrdersByShopIDUseCase_Execute_Query(t *testing.T) {
	shopID := 1

	t.Run("when first page then returns orders totalCount and pagination info", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := defaultFilters() // LastID is nil = first page

		expectedOrders := newTestOrders()
		expectedCount := 10

		shopServiceMock := mocks.NewShopService(t)
		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedCount, nil)
		orderServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedOrders, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(expectedOrders, models.DefaultLimit, models.SortByCreatedAt).
			Return(expectedOrders, "", false)

		useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

		// Act
		orders, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedOrders, orders)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		require.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})

	t.Run("when subsequent page with LastID then returns orders without totalCount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		lastID := 5
		filters := models.OrderFilters{
			LastID: &lastID,
		}

		expectedOrders := newTestOrders()

		shopServiceMock := mocks.NewShopService(t)
		orderServiceMock := mocks.NewOrderService(t)
		// Only GetAllByShopIDWithFilters should be called, NOT CountByShopIDWithFilters
		orderServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedOrders, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(expectedOrders, models.DefaultLimit, models.SortByCreatedAt).
			Return(expectedOrders, "", false)

		useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

		// Act
		orders, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedOrders, orders)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when orderService returns error then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := defaultFilters()
		expectedError := stdErrors.New("database connection error")

		shopServiceMock := mocks.NewShopService(t)
		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(0, nil)
		orderServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(nil, expectedError)

		paginationServiceMock := mocks.NewPaginationService[*models.Order](t)

		useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

		// Act
		orders, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, orders)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

// =============================================================================
// GetAllOrdersByShopIDUseCase Pagination Tests
// =============================================================================

func TestGetAllOrdersByShopIDUseCase_Execute_Pagination(t *testing.T) {
	shopID := 1

	t.Run("when has more pages then returns hasMore true and nextCursor", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := defaultFilters()

		// Simulate LIMIT+1 strategy: 3 orders returned when limit is 20
		// PaginationService determines hasMore and trims
		allOrders := []*models.Order{
			{ID: 1, OrderNumber: "ORD-001", Total: 100.00, CreatedAt: time.Now()},
			{ID: 2, OrderNumber: "ORD-002", Total: 200.00, CreatedAt: time.Now()},
			{ID: 3, OrderNumber: "ORD-003", Total: 300.00, CreatedAt: time.Now()},
		}
		trimmedOrders := allOrders[:2]
		expectedCursor := "eyJpZCI6Miwic29ydF92YWx1ZSI6IjIwMjUtMDEtMDFUMDA6MDA6MDBaIn0="

		shopServiceMock := mocks.NewShopService(t)
		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(50, nil)
		orderServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(allOrders, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Order](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(allOrders, models.DefaultLimit, models.SortByCreatedAt).
			Return(trimmedOrders, expectedCursor, true)

		useCase := NewGetAllOrdersByShopIDUseCase(shopServiceMock, orderServiceMock, paginationServiceMock)

		// Act
		orders, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, trimmedOrders, orders)
		assert.Equal(t, expectedCursor, nextCursor)
		assert.True(t, hasMore)
		require.NotNil(t, totalCount)
		assert.Equal(t, 50, *totalCount)
	})
}
