package claims

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromContext(t *testing.T) {
	t.Run("when user_id exists in context then returns it", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyUserID, 42)

		// Act
		userID := GetUserIDFromContext(ctx)

		// Assert
		assert.Equal(t, 42, userID)
	})

	t.Run("when user_id is missing then returns 0", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		userID := GetUserIDFromContext(ctx)

		// Assert
		assert.Equal(t, 0, userID)
	})

	t.Run("when user_id is wrong type then returns 0", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyUserID, "not-an-int")

		// Act
		userID := GetUserIDFromContext(ctx)

		// Assert
		assert.Equal(t, 0, userID)
	})
}

func TestGetEmailFromContext(t *testing.T) {
	t.Run("when email exists in context then returns it", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyEmail, "test@example.com")

		// Act
		email := GetEmailFromContext(ctx)

		// Assert
		assert.Equal(t, "test@example.com", email)
	})

	t.Run("when email is missing then returns empty string", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		email := GetEmailFromContext(ctx)

		// Assert
		assert.Equal(t, "", email)
	})

	t.Run("when email is wrong type then returns empty string", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyEmail, 123)

		// Act
		email := GetEmailFromContext(ctx)

		// Assert
		assert.Equal(t, "", email)
	})
}

func TestGetShopIDsFromContext(t *testing.T) {
	t.Run("when shop_ids exist in context then returns them", func(t *testing.T) {
		// Arrange
		shopIDs := []int{1, 2, 3}
		ctx := context.WithValue(context.Background(), KeyShopIDs, shopIDs)

		// Act
		result := GetShopIDsFromContext(ctx)

		// Assert
		assert.Equal(t, shopIDs, result)
	})

	t.Run("when shop_ids is missing then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		result := GetShopIDsFromContext(ctx)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when shop_ids is wrong type then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyShopIDs, "not-a-slice")

		// Act
		result := GetShopIDsFromContext(ctx)

		// Assert
		assert.Nil(t, result)
	})
}

func TestTokenClaims_ShopIDs(t *testing.T) {
	t.Run("when ShopRoles has multiple entries then returns all shop IDs", func(t *testing.T) {
		// Arrange
		tc := &TokenClaims{
			UserID: 1,
			Email:  "test@example.com",
			ShopRoles: []ShopRole{
				{ShopID: 10, Role: "owner"},
				{ShopID: 20, Role: "admin"},
				{ShopID: 30, Role: "staff"},
			},
		}

		// Act
		ids := tc.ShopIDs()

		// Assert
		assert.Equal(t, []int{10, 20, 30}, ids)
	})

	t.Run("when ShopRoles is empty then returns empty slice", func(t *testing.T) {
		// Arrange
		tc := &TokenClaims{
			UserID:    1,
			Email:     "test@example.com",
			ShopRoles: []ShopRole{},
		}

		// Act
		ids := tc.ShopIDs()

		// Assert
		assert.Empty(t, ids)
		assert.NotNil(t, ids)
	})

	t.Run("when ShopRoles has single entry then returns single ID", func(t *testing.T) {
		// Arrange
		tc := &TokenClaims{
			UserID:    1,
			Email:     "test@example.com",
			ShopRoles: []ShopRole{{ShopID: 42, Role: "owner"}},
		}

		// Act
		ids := tc.ShopIDs()

		// Assert
		assert.Equal(t, []int{42}, ids)
	})
}

func TestGetShopRolesFromContext(t *testing.T) {
	t.Run("when shop_roles exist in context then returns them", func(t *testing.T) {
		// Arrange
		shopRoles := []ShopRole{
			{ShopID: 1, Role: "owner", Permissions: []string{"manage_staff"}},
			{ShopID: 2, Role: "admin", Permissions: []string{"manage_products"}},
		}
		ctx := context.WithValue(context.Background(), KeyShopRoles, shopRoles)

		// Act
		result := GetShopRolesFromContext(ctx)

		// Assert
		assert.Equal(t, shopRoles, result)
	})

	t.Run("when shop_roles is missing then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		result := GetShopRolesFromContext(ctx)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when shop_roles is wrong type then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyShopRoles, "not-a-slice")

		// Act
		result := GetShopRolesFromContext(ctx)

		// Assert
		assert.Nil(t, result)
	})
}

func TestGetRoleForShop(t *testing.T) {
	t.Run("when user has role for shop then returns role", func(t *testing.T) {
		// Arrange
		shopRoles := []ShopRole{
			{ShopID: 1, Role: "owner"},
			{ShopID: 2, Role: "admin"},
			{ShopID: 3, Role: "staff"},
		}
		ctx := context.WithValue(context.Background(), KeyShopRoles, shopRoles)

		// Act
		role := GetRoleForShop(ctx, 2)

		// Assert
		assert.Equal(t, "admin", role)
	})

	t.Run("when user has no role for shop then returns empty string", func(t *testing.T) {
		// Arrange
		shopRoles := []ShopRole{
			{ShopID: 1, Role: "owner"},
		}
		ctx := context.WithValue(context.Background(), KeyShopRoles, shopRoles)

		// Act
		role := GetRoleForShop(ctx, 99)

		// Assert
		assert.Equal(t, "", role)
	})

	t.Run("when shop_roles not in context then returns empty string", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		role := GetRoleForShop(ctx, 1)

		// Assert
		assert.Equal(t, "", role)
	})

	t.Run("when shop_roles is empty then returns empty string", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyShopRoles, []ShopRole{})

		// Act
		role := GetRoleForShop(ctx, 1)

		// Assert
		assert.Equal(t, "", role)
	})
}

func TestGetFirstShopIDFromContext(t *testing.T) {
	t.Run("when shop_ids exist then returns first one", func(t *testing.T) {
		// Arrange
		shopIDs := []int{5, 10, 15}
		ctx := context.WithValue(context.Background(), KeyShopIDs, shopIDs)

		// Act
		result := GetFirstShopIDFromContext(ctx)

		// Assert
		assert.Equal(t, 5, result)
	})

	t.Run("when shop_ids is empty then returns 0", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), KeyShopIDs, []int{})

		// Act
		result := GetFirstShopIDFromContext(ctx)

		// Assert
		assert.Equal(t, 0, result)
	})

	t.Run("when shop_ids is missing then returns 0", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		result := GetFirstShopIDFromContext(ctx)

		// Assert
		assert.Equal(t, 0, result)
	})
}
