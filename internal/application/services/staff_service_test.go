package services

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func init() {
	logs.Init()
}

// ============================================================================
// Helper functions
// ============================================================================

func newStaffService(t *testing.T) (*StaffServiceImpl, *mocks.StaffRepository, *mocks.RoleRepository, *mocks.AssetService, *mocks.AuthService) {
	staffRepo := mocks.NewStaffRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	assetSvc := mocks.NewAssetService(t)
	authSvc := mocks.NewAuthService(t)
	svc := NewStaffService(staffRepo, roleRepo, assetSvc, authSvc)
	return svc, staffRepo, roleRepo, assetSvc, authSvc
}

func validUser() *models.User {
	return &models.User{
		Name:     "John",
		LastName: "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
}

func ownerStaff() *models.Staff {
	return &models.Staff{
		ID:       1,
		ShopID:   1,
		IsActive: true,
		User:     &models.User{ID: 10, Name: "Owner"},
		Roles:    []*models.Role{{ID: 1, Name: "owner"}},
	}
}

func regularStaff() *models.Staff {
	return &models.Staff{
		ID:       2,
		ShopID:   1,
		IsActive: true,
		User: &models.User{
			ID:           20,
			Name:         "Staff",
			ProfileImage: &models.Image{ID: 1, URL: "https://old.com/img.png", StorageRef: "old_ref"},
		},
		Roles: []*models.Role{{ID: 2, Name: "admin"}},
	}
}

// ============================================================================
// Create
// ============================================================================

func TestStaffService_Create(t *testing.T) {
	t.Run("when create is successful with image then returns staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, authSvc := newStaffService(t)
		user := validUser()
		shopID := 1
		roleID := 2
		imageBuffer := []byte("image-data")
		uploadedImage := &models.Image{URL: "https://cdn.com/img.png", StorageRef: "shop_1/staff/img.png"}
		expectedStaff := &models.Staff{ID: 1, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("hashed_password", nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().CreateWithUser(ctx, shopID, user, roleID, uploadedImage).Return(expectedStaff, nil)

		// Act
		staff, err := svc.Create(ctx, shopID, user, roleID, imageBuffer)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedStaff, staff)
	})

	t.Run("when create is successful without image then returns staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, authSvc := newStaffService(t)
		user := validUser()
		shopID := 1
		roleID := 2
		expectedStaff := &models.Staff{ID: 1, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("hashed_password", nil)
		staffRepo.EXPECT().CreateWithUser(ctx, shopID, user, roleID, (*models.Image)(nil)).Return(expectedStaff, nil)

		// Act
		staff, err := svc.Create(ctx, shopID, user, roleID, nil)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedStaff, staff)
	})

	t.Run("when roleID is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, _, _, _ := newStaffService(t)
		user := validUser()

		// Act
		staff, err := svc.Create(ctx, 1, user, 0, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.RoleNotFound, validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when role is owner then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		roleID := 1

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "owner"}, nil)

		// Act
		staff, err := svc.Create(ctx, 1, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OwnerRoleNotAllowed, validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when role not found in DB then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		roleID := 999

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(nil, stdErrors.New("not found"))

		// Act
		staff, err := svc.Create(ctx, 1, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.RoleNotFound, validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when user name is empty then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Name = ""
		roleID := 2

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)

		// Act
		staff, err := svc.Create(ctx, 1, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "user_name_is_required", validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when user last name is empty then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.LastName = ""
		roleID := 2

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)

		// Act
		staff, err := svc.Create(ctx, 1, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "user_last_name_is_required", validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when user email is empty then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Email = ""
		roleID := 2

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)

		// Act
		staff, err := svc.Create(ctx, 1, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "user_email_is_required", validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when password is too short then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Password = "short"
		roleID := 2

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)

		// Act
		staff, err := svc.Create(ctx, 1, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "password_must_be_at_least_6_characters", validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when HashPassword fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, authSvc := newStaffService(t)
		user := validUser()
		roleID := 2
		expectedError := stdErrors.New("hash error")

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("", expectedError)

		// Act
		staff, err := svc.Create(ctx, 1, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staff)
	})

	t.Run("when asset upload fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, assetSvc, authSvc := newStaffService(t)
		user := validUser()
		shopID := 1
		roleID := 2
		imageBuffer := []byte("image-data")
		expectedError := stdErrors.New("upload failed")

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("hashed_password", nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(nil, expectedError)

		// Act
		staff, err := svc.Create(ctx, shopID, user, roleID, imageBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staff)
	})

	t.Run("when CreateWithUser fails then cleans up uploaded image and returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, authSvc := newStaffService(t)
		user := validUser()
		shopID := 1
		roleID := 2
		imageBuffer := []byte("image-data")
		uploadedImage := &models.Image{URL: "https://cdn.com/img.png", StorageRef: "shop_1/staff/img.png"}
		expectedError := stdErrors.New("db error")

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("hashed_password", nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().CreateWithUser(ctx, shopID, user, roleID, uploadedImage).Return(nil, expectedError)
		assetSvc.EXPECT().Delete(ctx, "shop_1/staff/img.png").Return(nil)

		// Act
		staff, err := svc.Create(ctx, shopID, user, roleID, imageBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staff)
	})
}

// ============================================================================
// Update
// ============================================================================

func TestStaffService_Update(t *testing.T) {
	t.Run("when update is successful with image then returns staff and cleans old image", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, authSvc := newStaffService(t)
		user := validUser()
		user.Password = "" // no password change
		shopID := 1
		staffID := 2
		roleID := 2
		imageBuffer := []byte("new-image")
		existingStaff := regularStaff()
		uploadedImage := &models.Image{URL: "https://cdn.com/new.png", StorageRef: "new_ref"}
		updatedStaff := &models.Staff{ID: staffID, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, uploadedImage).Return(updatedStaff, nil)
		assetSvc.EXPECT().Delete(ctx, "old_ref").Return(nil)
		_ = authSvc // not called since password is empty

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, updatedStaff, staff)
	})

	t.Run("when update is successful without image then returns staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		existingStaff := regularStaff()
		updatedStaff := &models.Staff{ID: staffID, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, (*models.Image)(nil)).Return(updatedStaff, nil)

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, nil)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, updatedStaff, staff)
	})

	t.Run("when staff is owner then returns owner_cannot_be_modified error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 1
		roleID := 2

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(ownerStaff(), nil)

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OwnerCannotBeModified, validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when validateRole fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, _, _, _ := newStaffService(t)
		user := validUser()

		// Act
		staff, err := svc.Update(ctx, 2, 1, user, 0, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.RoleNotFound, validationErr.Message)
		assert.Nil(t, staff)
	})

	t.Run("when password is provided and HashPassword fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, authSvc := newStaffService(t)
		user := validUser()
		shopID := 1
		staffID := 2
		roleID := 2
		expectedError := stdErrors.New("hash error")
		existingStaff := regularStaff()

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("", expectedError)

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staff)
	})

	t.Run("when password is empty then no hashing is performed", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		existingStaff := regularStaff()
		updatedStaff := &models.Staff{ID: staffID, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, (*models.Image)(nil)).Return(updatedStaff, nil)
		// authSvc.HashPassword is NOT expected to be called

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, nil)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, updatedStaff, staff)
		assert.Equal(t, "", user.Password)
	})
}

// ============================================================================
// Delete
// ============================================================================

func TestStaffService_Delete(t *testing.T) {
	t.Run("when delete is successful then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		staffID := 2
		existing := regularStaff()

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existing, nil)
		staffRepo.EXPECT().Delete(ctx, staffID, shopID).Return(nil)

		// Act
		err := svc.Delete(ctx, staffID, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when staff not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		staffID := 999
		expectedError := stdErrors.New(errors.StaffNotFound)

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(nil, expectedError)

		// Act
		err := svc.Delete(ctx, staffID, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when staff is owner then returns owner_cannot_be_deleted error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		staffID := 1

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(ownerStaff(), nil)

		// Act
		err := svc.Delete(ctx, staffID, shopID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OwnerCannotBeDeleted, validationErr.Message)
	})
}

// ============================================================================
// ToggleStatus
// ============================================================================

func TestStaffService_ToggleStatus(t *testing.T) {
	t.Run("when toggle is successful then returns updated staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		staffID := 2
		existing := regularStaff()
		toggledStaff := &models.Staff{ID: staffID, ShopID: shopID, IsActive: false}

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existing, nil)
		staffRepo.EXPECT().ToggleStatus(ctx, staffID, shopID).Return(toggledStaff, nil)

		// Act
		staff, err := svc.ToggleStatus(ctx, staffID, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, toggledStaff, staff)
	})

	t.Run("when staff is owner then returns owner_status_cannot_be_changed error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		staffID := 1

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(ownerStaff(), nil)

		// Act
		staff, err := svc.ToggleStatus(ctx, staffID, shopID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OwnerStatusCannotChange, validationErr.Message)
		assert.Nil(t, staff)
	})
}

// ============================================================================
// GetShopRolesByUserID
// ============================================================================

// ============================================================================
// GetAllByShopIDWithFilters
// ============================================================================

func TestStaffService_GetAllByShopIDWithFilters(t *testing.T) {
	t.Run("when successful then returns staff list", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		filters := models.StaffFilters{Limit: 20}
		expectedStaff := []*models.Staff{regularStaff()}

		staffRepo.EXPECT().GetAllByShopIDWithFilters(ctx, shopID, filters).Return(expectedStaff, nil)

		// Act
		staff, err := svc.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedStaff, staff)
	})

	t.Run("when repository fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		filters := models.StaffFilters{Limit: 20}
		expectedError := stdErrors.New("db error")

		staffRepo.EXPECT().GetAllByShopIDWithFilters(ctx, shopID, filters).Return(nil, expectedError)

		// Act
		staff, err := svc.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staff)
	})
}

// ============================================================================
// CountByShopIDWithFilters
// ============================================================================

func TestStaffService_CountByShopIDWithFilters(t *testing.T) {
	t.Run("when successful then returns count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		filters := models.StaffFilters{Limit: 20}

		staffRepo.EXPECT().CountByShopIDWithFilters(ctx, shopID, filters).Return(5, nil)

		// Act
		count, err := svc.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("when repository fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		filters := models.StaffFilters{Limit: 20}
		expectedError := stdErrors.New("db error")

		staffRepo.EXPECT().CountByShopIDWithFilters(ctx, shopID, filters).Return(0, expectedError)

		// Act
		count, err := svc.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, 0, count)
	})
}

// ============================================================================
// GetByID
// ============================================================================

func TestStaffService_GetByID(t *testing.T) {
	t.Run("when successful then returns staff", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		staffID := 2
		shopID := 1
		expectedStaff := regularStaff()

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(expectedStaff, nil)

		// Act
		staff, err := svc.GetByID(ctx, staffID, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedStaff, staff)
	})

	t.Run("when repository fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		staffID := 999
		shopID := 1
		expectedError := stdErrors.New("db error")

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(nil, expectedError)

		// Act
		staff, err := svc.GetByID(ctx, staffID, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staff)
	})
}

// ============================================================================
// GetShopRolesByUserID
// ============================================================================

func TestStaffService_GetShopRolesByUserID(t *testing.T) {
	t.Run("when successful then returns shop roles", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		userID := 10
		expectedRoles := []claims.ShopRole{
			{ShopID: 1, Role: "admin"},
			{ShopID: 2, Role: "staff"},
		}

		staffRepo.EXPECT().GetShopRolesByUserID(ctx, userID).Return(expectedRoles, nil)

		// Act
		roles, err := svc.GetShopRolesByUserID(ctx, userID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedRoles, roles)
	})

	t.Run("when repository fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		userID := 10
		expectedError := stdErrors.New("db error")

		staffRepo.EXPECT().GetShopRolesByUserID(ctx, userID).Return(nil, expectedError)

		// Act
		roles, err := svc.GetShopRolesByUserID(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, roles)
	})
}

// ============================================================================
// Update — additional branch coverage
// ============================================================================

func TestStaffService_Update_CleanupBranches(t *testing.T) {
	t.Run("when update repo fails and image cleanup also fails then returns repo error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		imageBuffer := []byte("new-image")
		existingStaff := regularStaff()
		uploadedImage := &models.Image{URL: "https://cdn.com/new.png", StorageRef: "new_ref"}
		repoError := stdErrors.New("db error")
		cleanupError := stdErrors.New("cleanup failed")

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, uploadedImage).Return(nil, repoError)
		assetSvc.EXPECT().Delete(ctx, "new_ref").Return(cleanupError)

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoError, err)
		assert.Nil(t, staff)
	})

	t.Run("when update succeeds but old image cloud delete fails then returns staff without error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		imageBuffer := []byte("new-image")
		existingStaff := regularStaff()
		uploadedImage := &models.Image{URL: "https://cdn.com/new.png", StorageRef: "new_ref"}
		updatedStaff := &models.Staff{ID: staffID, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, uploadedImage).Return(updatedStaff, nil)
		assetSvc.EXPECT().Delete(ctx, "old_ref").Return(stdErrors.New("cloud delete failed"))

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, updatedStaff, staff)
	})

	t.Run("when GetByID fails during update then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		dbError := stdErrors.New("db connection error")

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(nil, dbError)

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, staff)
	})

	t.Run("when image upload fails during update then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		imageBuffer := []byte("new-image")
		existingStaff := regularStaff()
		uploadError := stdErrors.New("upload failed")

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(nil, uploadError)

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uploadError, err)
		assert.Nil(t, staff)
	})
}

// ============================================================================
// ToggleStatus — additional branch coverage
// ============================================================================

func TestStaffService_ToggleStatus_Additional(t *testing.T) {
	t.Run("when GetByID returns a generic DB error then returns that error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, _, _, _ := newStaffService(t)
		shopID := 1
		staffID := 2
		expectedError := stdErrors.New("connection refused")

		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(nil, expectedError)

		// Act
		staff, err := svc.ToggleStatus(ctx, staffID, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, staff)
	})
}

// ============================================================================
// cleanupAsset — additional branch coverage
// ============================================================================

func TestStaffService_CleanupAsset_NilAndEmptyBranches(t *testing.T) {
	t.Run("when image is nil then does nothing", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, authSvc := newStaffService(t)
		user := validUser()
		shopID := 1
		roleID := 2

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("hashed_password", nil)
		// CreateWithUser fails so cleanupAsset is called, but image is nil (no upload)
		staffRepo.EXPECT().CreateWithUser(ctx, shopID, user, roleID, (*models.Image)(nil)).Return(nil, stdErrors.New("db error"))
		// assetSvc.Delete should NOT be called since image is nil

		// Act
		staff, err := svc.Create(ctx, shopID, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, staff)
	})

	t.Run("when image has empty StorageRef then does nothing", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, authSvc := newStaffService(t)
		user := validUser()
		shopID := 1
		roleID := 2
		imageBuffer := []byte("image-data")
		uploadedImage := &models.Image{URL: "https://cdn.com/img.png", StorageRef: ""}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("hashed_password", nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().CreateWithUser(ctx, shopID, user, roleID, uploadedImage).Return(nil, stdErrors.New("db error"))
		// Delete should NOT be called since StorageRef is empty

		// Act
		staff, err := svc.Create(ctx, shopID, user, roleID, imageBuffer)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, staff)
	})
}

// ============================================================================
// cleanupOldImage — additional branch coverage
// ============================================================================

func TestStaffService_CleanupOldImage_NilBranches(t *testing.T) {
	t.Run("when existingStaff.User is nil then skips old image cleanup", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		imageBuffer := []byte("new-image")
		existingStaff := &models.Staff{
			ID:     staffID,
			ShopID: shopID,
			User:   nil, // User is nil
			Roles:  []*models.Role{{ID: 2, Name: "admin"}},
		}
		uploadedImage := &models.Image{URL: "https://cdn.com/new.png", StorageRef: "new_ref"}
		updatedStaff := &models.Staff{ID: staffID, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, uploadedImage).Return(updatedStaff, nil)
		// Delete should NOT be called since User is nil

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, updatedStaff, staff)
	})

	t.Run("when existingStaff.User.ProfileImage is nil then skips old image cleanup", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		imageBuffer := []byte("new-image")
		existingStaff := &models.Staff{
			ID:     staffID,
			ShopID: shopID,
			User:   &models.User{ID: 20, Name: "Staff", ProfileImage: nil},
			Roles:  []*models.Role{{ID: 2, Name: "admin"}},
		}
		uploadedImage := &models.Image{URL: "https://cdn.com/new.png", StorageRef: "new_ref"}
		updatedStaff := &models.Staff{ID: staffID, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, uploadedImage).Return(updatedStaff, nil)
		// Delete should NOT be called since ProfileImage is nil

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, updatedStaff, staff)
	})

	t.Run("when existingStaff.User.ProfileImage.StorageRef is empty then skips old image cleanup", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, assetSvc, _ := newStaffService(t)
		user := validUser()
		user.Password = ""
		shopID := 1
		staffID := 2
		roleID := 2
		imageBuffer := []byte("new-image")
		existingStaff := &models.Staff{
			ID:     staffID,
			ShopID: shopID,
			User:   &models.User{ID: 20, Name: "Staff", ProfileImage: &models.Image{ID: 1, URL: "https://old.com/img.png", StorageRef: ""}},
			Roles:  []*models.Role{{ID: 2, Name: "admin"}},
		}
		uploadedImage := &models.Image{URL: "https://cdn.com/new.png", StorageRef: "new_ref"}
		updatedStaff := &models.Staff{ID: staffID, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		staffRepo.EXPECT().GetByID(ctx, staffID, shopID).Return(existingStaff, nil)
		assetSvc.EXPECT().Upload(ctx, imageBuffer, "shop_1/staff").Return(uploadedImage, nil)
		staffRepo.EXPECT().Update(ctx, staffID, shopID, user, roleID, uploadedImage).Return(updatedStaff, nil)
		// Delete should NOT be called since StorageRef is empty

		// Act
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, imageBuffer)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, updatedStaff, staff)
	})
}

// ============================================================================
// validateUserFields — additional branch coverage
// ============================================================================

func TestStaffService_ValidateUserFields_Additional(t *testing.T) {
	t.Run("when requirePassword is true and password is valid then passes validation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, staffRepo, roleRepo, _, authSvc := newStaffService(t)
		user := validUser() // password = "password123" (>= 6 chars)
		shopID := 1
		roleID := 2
		expectedStaff := &models.Staff{ID: 1, ShopID: shopID}

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		authSvc.EXPECT().HashPassword(ctx, "password123").Return("hashed_password", nil)
		staffRepo.EXPECT().CreateWithUser(ctx, shopID, user, roleID, (*models.Image)(nil)).Return(expectedStaff, nil)

		// Act — Create uses requirePassword=true
		staff, err := svc.Create(ctx, shopID, user, roleID, nil)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedStaff, staff)
	})

	t.Run("when requirePassword is false and password is short then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		svc, _, roleRepo, _, _ := newStaffService(t)
		user := validUser()
		user.Password = "short" // < 6 chars, requirePassword=false path
		staffID := 2
		shopID := 1
		roleID := 2

		roleRepo.EXPECT().GetByID(ctx, roleID).Return(&models.Role{ID: roleID, Name: "admin"}, nil)
		// validateUserFields is called before GetByID, so GetByID is never reached

		// Act — Update uses requirePassword=false
		staff, err := svc.Update(ctx, staffID, shopID, user, roleID, nil)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "password_must_be_at_least_6_characters", validationErr.Message)
		assert.Nil(t, staff)
	})
}
