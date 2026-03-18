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
