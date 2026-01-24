package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Shop.Validate Tests
// =============================================================================

func TestShop_Validate(t *testing.T) {
	t.Run("when primary_color is nil then returns no error", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: nil,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when primary_color is empty string then returns no error", func(t *testing.T) {
		// Arrange
		emptyColor := ""
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &emptyColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when primary_color is valid hex format then returns no error", func(t *testing.T) {
		// Arrange
		validColor := "#FF5733"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &validColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when primary_color is valid lowercase hex then returns no error", func(t *testing.T) {
		// Arrange
		validColor := "#ff5733"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &validColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when primary_color is valid mixed case hex then returns no error", func(t *testing.T) {
		// Arrange
		validColor := "#Ff5733"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &validColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when primary_color is black then returns no error", func(t *testing.T) {
		// Arrange
		blackColor := "#000000"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &blackColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when primary_color is white then returns no error", func(t *testing.T) {
		// Arrange
		whiteColor := "#FFFFFF"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &whiteColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when primary_color has invalid hex format without hash then returns validation error", func(t *testing.T) {
		// Arrange
		invalidColor := "FF5733"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &invalidColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidPrimaryColorFormat, validationErr.Message)
	})

	t.Run("when primary_color has only 3 hex digits then returns validation error", func(t *testing.T) {
		// Arrange
		shortColor := "#FFF"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &shortColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidPrimaryColorFormat, validationErr.Message)
	})

	t.Run("when primary_color has 7 hex digits then returns validation error", func(t *testing.T) {
		// Arrange
		longColor := "#FFFFFFF"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &longColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidPrimaryColorFormat, validationErr.Message)
	})

	t.Run("when primary_color has invalid characters then returns validation error", func(t *testing.T) {
		// Arrange
		invalidColor := "#GG5733"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &invalidColor,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidPrimaryColorFormat, validationErr.Message)
	})

	t.Run("when primary_color is color name then returns validation error", func(t *testing.T) {
		// Arrange
		colorName := "red"
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &colorName,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidPrimaryColorFormat, validationErr.Message)
	})

	t.Run("when primary_color has spaces then returns validation error", func(t *testing.T) {
		// Arrange
		colorWithSpaces := " #FF5733 "
		shop := &Shop{
			ID:           1,
			Name:         "Test Shop",
			PrimaryColor: &colorWithSpaces,
		}

		// Act
		err := shop.Validate()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidPrimaryColorFormat, validationErr.Message)
	})
}
