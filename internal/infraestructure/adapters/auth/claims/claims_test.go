package claims

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

func TestNewTokenClaims(t *testing.T) {
	t.Run("when claims are valid then returns TokenClaims", func(t *testing.T) {
		// Arrange
		user := models.User{ID: 1, Email: "test@example.com"}
		userJSON, _ := json.Marshal(user)
		claims := map[string]interface{}{
			"user":     string(userJSON),
			"shop_ids": []interface{}{float64(1), float64(2)},
		}

		// Act
		tokenClaims, err := NewTokenClaims(claims)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tokenClaims)
		assert.Equal(t, 1, tokenClaims.UserID)
		assert.Equal(t, "test@example.com", tokenClaims.Email)
		assert.Equal(t, []int{1, 2}, tokenClaims.ShopIDs)
	})

	t.Run("when user claim is missing then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		claims := map[string]interface{}{
			"shop_ids": []interface{}{float64(1)},
		}

		// Act
		tokenClaims, err := NewTokenClaims(claims)

		// Assert
		assert.Nil(t, tokenClaims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.TokenInvalid, authErr.Message)
	})

	t.Run("when user claim is not a string then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		claims := map[string]interface{}{
			"user":     123, // Not a string
			"shop_ids": []interface{}{float64(1)},
		}

		// Act
		tokenClaims, err := NewTokenClaims(claims)

		// Assert
		assert.Nil(t, tokenClaims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.TokenInvalid, authErr.Message)
	})

	t.Run("when user claim is invalid JSON then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		claims := map[string]interface{}{
			"user":     "invalid-json",
			"shop_ids": []interface{}{float64(1)},
		}

		// Act
		tokenClaims, err := NewTokenClaims(claims)

		// Assert
		assert.Nil(t, tokenClaims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.TokenInvalid, authErr.Message)
	})

	t.Run("when shop_ids is missing then returns empty slice", func(t *testing.T) {
		// Arrange
		user := models.User{ID: 1, Email: "test@example.com"}
		userJSON, _ := json.Marshal(user)
		claims := map[string]interface{}{
			"user": string(userJSON),
		}

		// Act
		tokenClaims, err := NewTokenClaims(claims)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tokenClaims)
		assert.Empty(t, tokenClaims.ShopIDs)
	})

	t.Run("when shop_ids contains non-numeric values then skips them", func(t *testing.T) {
		// Arrange
		user := models.User{ID: 1, Email: "test@example.com"}
		userJSON, _ := json.Marshal(user)
		claims := map[string]interface{}{
			"user":     string(userJSON),
			"shop_ids": []interface{}{float64(1), "invalid", float64(3)},
		}

		// Act
		tokenClaims, err := NewTokenClaims(claims)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tokenClaims)
		assert.Equal(t, []int{1, 3}, tokenClaims.ShopIDs)
	})

	t.Run("when shop_ids is not a slice then returns empty slice", func(t *testing.T) {
		// Arrange
		user := models.User{ID: 1, Email: "test@example.com"}
		userJSON, _ := json.Marshal(user)
		claims := map[string]interface{}{
			"user":     string(userJSON),
			"shop_ids": "not-a-slice",
		}

		// Act
		tokenClaims, err := NewTokenClaims(claims)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tokenClaims)
		assert.Empty(t, tokenClaims.ShopIDs)
	})
}

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
