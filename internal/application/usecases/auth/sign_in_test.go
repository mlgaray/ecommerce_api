package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/mocks"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

func TestSignInUseCase_Execute(t *testing.T) {
	t.Run("when sign in with valid credentials then returns token successfully", func(t *testing.T) {
		ctx := context.Background()
		inputUser := &models.User{Email: "user@example.com", Password: "password123"}
		storedUser := &models.User{ID: 1, Email: "user@example.com", Password: "hashedpassword"}
		shopRoles := []claims.ShopRole{{ShopID: 10, Role: "owner"}}
		ownerPerms := []string{"create_product", "manage_staff"}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)
		staffServiceMock := new(mocks.StaffService)
		permServiceMock := new(mocks.PermissionService)

		userServiceMock.EXPECT().GetByEmail(ctx, "user@example.com").Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, "hashedpassword").Return(inputUser, nil)
		staffServiceMock.EXPECT().GetShopRolesByUserID(ctx, storedUser.ID).Return(shopRoles, nil)
		permServiceMock.EXPECT().GetPermissions("owner").Return(ownerPerms)

		// After permissions are resolved, shopRoles will have permissions attached
		expectedShopRoles := []claims.ShopRole{{ShopID: 10, Role: "owner", Permissions: ownerPerms}}
		tokenServiceMock.EXPECT().Generate(ctx, storedUser, expectedShopRoles).Return("jwt.token.here", nil)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, staffServiceMock, permServiceMock)
		token, err := useCase.Execute(ctx, inputUser)

		assert.NoError(t, err)
		assert.Equal(t, "jwt.token.here", token)
	})

	t.Run("when user not found then returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.New("user_not_found")
		inputUser := &models.User{Email: "notfound@example.com", Password: "password123"}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)
		staffServiceMock := new(mocks.StaffService)
		permServiceMock := new(mocks.PermissionService)

		userServiceMock.EXPECT().GetByEmail(ctx, "notfound@example.com").Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, staffServiceMock, permServiceMock)
		token, err := useCase.Execute(ctx, inputUser)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})

	t.Run("when credentials are invalid then returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.New("invalid credentials")
		inputUser := &models.User{Email: "user@example.com", Password: "wrongpassword"}
		storedUser := &models.User{ID: 1, Email: "user@example.com", Password: "hashedpassword"}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)
		staffServiceMock := new(mocks.StaffService)
		permServiceMock := new(mocks.PermissionService)

		userServiceMock.EXPECT().GetByEmail(ctx, "user@example.com").Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, "hashedpassword").Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, staffServiceMock, permServiceMock)
		token, err := useCase.Execute(ctx, inputUser)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})

	t.Run("when token generation fails then returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.New("token generation failed")
		inputUser := &models.User{Email: "user@example.com", Password: "password123"}
		storedUser := &models.User{ID: 1, Email: "user@example.com", Password: "hashedpassword"}
		shopRoles := []claims.ShopRole{{ShopID: 10, Role: "owner"}}
		ownerPerms := []string{"create_product"}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)
		staffServiceMock := new(mocks.StaffService)
		permServiceMock := new(mocks.PermissionService)

		userServiceMock.EXPECT().GetByEmail(ctx, "user@example.com").Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, "hashedpassword").Return(inputUser, nil)
		staffServiceMock.EXPECT().GetShopRolesByUserID(ctx, storedUser.ID).Return(shopRoles, nil)
		permServiceMock.EXPECT().GetPermissions("owner").Return(ownerPerms)
		tokenServiceMock.EXPECT().Generate(ctx, storedUser, []claims.ShopRole{{ShopID: 10, Role: "owner", Permissions: ownerPerms}}).Return("", expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, staffServiceMock, permServiceMock)
		token, err := useCase.Execute(ctx, inputUser)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})

	t.Run("when get staff roles fails then returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.New("database error")
		inputUser := &models.User{Email: "user@example.com", Password: "password123"}
		storedUser := &models.User{ID: 1, Email: "user@example.com", Password: "hashedpassword"}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)
		staffServiceMock := new(mocks.StaffService)
		permServiceMock := new(mocks.PermissionService)

		userServiceMock.EXPECT().GetByEmail(ctx, "user@example.com").Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, "hashedpassword").Return(inputUser, nil)
		staffServiceMock.EXPECT().GetShopRolesByUserID(ctx, storedUser.ID).Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, staffServiceMock, permServiceMock)
		token, err := useCase.Execute(ctx, inputUser)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})
}
