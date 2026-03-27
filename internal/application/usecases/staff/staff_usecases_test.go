package staff

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestUser() *models.User {
	return &models.User{
		Name:     "John",
		LastName: "Doe",
		Email:    "john@example.com",
		Phone:    "1234567890",
	}
}

func newTestStaff(id int) *models.Staff {
	return &models.Staff{
		ID:        id,
		ShopID:    1,
		IsActive:  true,
		User:      newTestUser(),
		CreatedAt: time.Now(),
	}
}

func newTestStaffList() []*models.Staff {
	return []*models.Staff{
		newTestStaff(1),
		newTestStaff(2),
	}
}

// mockPaginationService is a local mock for ports.PaginationService[*models.Staff]
// since the generic mock requires type instantiation.
type mockPaginationService struct {
	mock.Mock
}

func (m *mockPaginationService) ApplyPagination(items []*models.Staff, limit int, sortBy string) ([]*models.Staff, string, bool) {
	args := m.Called(items, limit, sortBy)
	staff, _ := args.Get(0).([]*models.Staff)
	return staff, args.String(1), args.Bool(2)
}

func newMockPaginationService(t *testing.T) *mockPaginationService {
	m := &mockPaginationService{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

// =============================================================================
// CreateStaffUseCase Tests
// =============================================================================

func TestCreateStaffUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns created staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		user := newTestUser()
		roleID := 2
		imageBuffer := []byte("fake-image")
		expected := newTestStaff(1)

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			Create(ctx, shopID, user, roleID, imageBuffer).
			Return(expected, nil)

		useCase := NewCreateStaffUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, user, roleID, imageBuffer)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service returns validation error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		user := newTestUser()
		roleID := 2
		imageBuffer := []byte("fake-image")
		expectedError := &errors.ValidationError{Message: errors.StaffAlreadyExists}

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			Create(ctx, shopID, user, roleID, imageBuffer).
			Return(nil, expectedError)

		useCase := NewCreateStaffUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, user, roleID, imageBuffer)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})
}

// =============================================================================
// GetStaffByIDUseCase Tests
// =============================================================================

func TestGetStaffByIDUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 1
		shopID := 1
		expected := newTestStaff(staffID)

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			GetByID(ctx, staffID, shopID).
			Return(expected, nil)

		useCase := NewGetStaffByIDUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, staffID, shopID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 99
		shopID := 1
		expectedError := &errors.RecordNotFoundError{Message: errors.StaffNotFound}

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			GetByID(ctx, staffID, shopID).
			Return(nil, expectedError)

		useCase := NewGetStaffByIDUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, staffID, shopID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})
}

// =============================================================================
// UpdateStaffUseCase Tests
// =============================================================================

func TestUpdateStaffUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns updated staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 1
		shopID := 1
		user := newTestUser()
		roleID := 2
		imageBuffer := []byte("new-image")
		expected := newTestStaff(staffID)

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			Update(ctx, staffID, shopID, user, roleID, imageBuffer).
			Return(expected, nil)

		useCase := NewUpdateStaffUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service returns not found error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 99
		shopID := 1
		user := newTestUser()
		roleID := 2
		imageBuffer := []byte("new-image")
		expectedError := &errors.RecordNotFoundError{Message: errors.StaffNotFound}

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			Update(ctx, staffID, shopID, user, roleID, imageBuffer).
			Return(nil, expectedError)

		useCase := NewUpdateStaffUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})
}

// =============================================================================
// DeleteStaffUseCase Tests
// =============================================================================

func TestDeleteStaffUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns nil error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 1
		shopID := 1

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			Delete(ctx, staffID, shopID).
			Return(nil)

		useCase := NewDeleteStaffUseCase(staffServiceMock)

		// Act
		err := useCase.Execute(ctx, staffID, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 99
		shopID := 1
		expectedError := &errors.RecordNotFoundError{Message: errors.StaffNotFound}

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			Delete(ctx, staffID, shopID).
			Return(expectedError)

		useCase := NewDeleteStaffUseCase(staffServiceMock)

		// Act
		err := useCase.Execute(ctx, staffID, shopID)

		// Assert
		require.Error(t, err)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})
}

// =============================================================================
// ToggleStaffStatusUseCase Tests
// =============================================================================

func TestToggleStaffStatusUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns toggled staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 1
		shopID := 1
		expected := newTestStaff(staffID)
		expected.IsActive = false

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			ToggleStatus(ctx, staffID, shopID).
			Return(expected, nil)

		useCase := NewToggleStaffStatusUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, staffID, shopID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.False(t, result.IsActive)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		staffID := 99
		shopID := 1
		expectedError := &errors.RecordNotFoundError{Message: errors.StaffNotFound}

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			ToggleStatus(ctx, staffID, shopID).
			Return(nil, expectedError)

		useCase := NewToggleStaffStatusUseCase(staffServiceMock)

		// Act
		result, err := useCase.Execute(ctx, staffID, shopID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})
}

// =============================================================================
// GetAllStaffByShopIDUseCase Tests
// =============================================================================

func TestGetAllStaffByShopIDUseCase_Execute_Validation(t *testing.T) {
	t.Run("when invalid sort field then returns ValidationError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.StaffFilters{
			SortBy: "invalid_field",
		}

		staffServiceMock := mocks.NewStaffService(t)
		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllStaffByShopIDUseCase(staffServiceMock, paginationServiceMock)

		// Act
		staffList, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.InvalidSortField, valErr.Message)
		assert.Nil(t, staffList)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

func TestGetAllStaffByShopIDUseCase_Execute_FirstPage(t *testing.T) {
	shopID := 1

	t.Run("when first page succeeds then returns staff with totalCount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.StaffFilters{} // LastID is nil = first page
		expectedStaff := newTestStaffList()
		expectedCount := 10

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedCount, nil)
		staffServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedStaff, nil)

		paginationServiceMock := newMockPaginationService(t)
		paginationServiceMock.On("ApplyPagination", expectedStaff, models.DefaultLimit, models.SortByCreatedAt).
			Return(expectedStaff, "", false)

		useCase := NewGetAllStaffByShopIDUseCase(staffServiceMock, paginationServiceMock)

		// Act
		staffList, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedStaff, staffList)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		require.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})

	t.Run("when count fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.StaffFilters{}
		expectedError := stdErrors.New("count query failed")

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(0, expectedError)
		staffServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(nil, nil)

		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllStaffByShopIDUseCase(staffServiceMock, paginationServiceMock)

		// Act
		staffList, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staffList)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when select fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.StaffFilters{}
		expectedError := stdErrors.New("select query failed")

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(5, nil)
		staffServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(nil, expectedError)

		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllStaffByShopIDUseCase(staffServiceMock, paginationServiceMock)

		// Act
		staffList, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staffList)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

func TestGetAllStaffByShopIDUseCase_Execute_SubsequentPage(t *testing.T) {
	shopID := 1

	t.Run("when subsequent page succeeds then returns staff without totalCount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		lastID := 5
		filters := models.StaffFilters{
			LastID: &lastID,
		}
		expectedStaff := newTestStaffList()

		staffServiceMock := mocks.NewStaffService(t)
		// Only GetAllByShopIDWithFilters should be called, NOT CountByShopIDWithFilters
		staffServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedStaff, nil)

		paginationServiceMock := newMockPaginationService(t)
		paginationServiceMock.On("ApplyPagination", expectedStaff, models.DefaultLimit, models.SortByCreatedAt).
			Return(expectedStaff, "", false)

		useCase := NewGetAllStaffByShopIDUseCase(staffServiceMock, paginationServiceMock)

		// Act
		staffList, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedStaff, staffList)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when subsequent page select fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		lastID := 5
		filters := models.StaffFilters{
			LastID: &lastID,
		}
		expectedError := stdErrors.New("database error")

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(nil, expectedError)

		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllStaffByShopIDUseCase(staffServiceMock, paginationServiceMock)

		// Act
		staffList, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staffList)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

func TestGetAllStaffByShopIDUseCase_Execute_Pagination(t *testing.T) {
	t.Run("when has more pages then returns hasMore true and nextCursor", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.StaffFilters{}
		expectedCount := 50

		// Simulate LIMIT+1 strategy: 3 staff returned
		allStaff := []*models.Staff{
			newTestStaff(1),
			newTestStaff(2),
			newTestStaff(3),
		}
		trimmedStaff := allStaff[:2]
		expectedCursor := "eyJpZCI6Miwic29ydF92YWx1ZSI6IjIwMjUtMDEtMDFUMDA6MDA6MDBaIn0="

		staffServiceMock := mocks.NewStaffService(t)
		staffServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedCount, nil)
		staffServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(allStaff, nil)

		paginationServiceMock := newMockPaginationService(t)
		paginationServiceMock.On("ApplyPagination", allStaff, models.DefaultLimit, models.SortByCreatedAt).
			Return(trimmedStaff, expectedCursor, true)

		useCase := NewGetAllStaffByShopIDUseCase(staffServiceMock, paginationServiceMock)

		// Act
		staffList, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, trimmedStaff, staffList)
		assert.Equal(t, expectedCursor, nextCursor)
		assert.True(t, hasMore)
		require.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})
}
