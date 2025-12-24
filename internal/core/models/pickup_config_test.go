package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidPickupConfig() *PickupConfig {
	return &PickupConfig{
		Address:      "Av. Corrientes 1234",
		City:         "Buenos Aires",
		Province:     "CABA",
		PostalCode:   "C1043AAZ",
		Instructions: "Tocar timbre 2B",
	}
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestPickupConfig_Validate(t *testing.T) {
	t.Run("when config is valid then returns nil", func(t *testing.T) {
		config := newValidPickupConfig()

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("when address is empty then returns error", func(t *testing.T) {
		config := newValidPickupConfig()
		config.Address = ""

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.PickupAddressIsRequired, validationErr.Message)
	})

	t.Run("when city is empty then returns error", func(t *testing.T) {
		config := newValidPickupConfig()
		config.City = ""

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.PickupCityIsRequired, validationErr.Message)
	})

	t.Run("when province is empty then validation passes", func(t *testing.T) {
		config := newValidPickupConfig()
		config.Province = "" // Province is optional

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("when postal_code is empty then validation passes", func(t *testing.T) {
		config := newValidPickupConfig()
		config.PostalCode = "" // PostalCode is optional

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("when instructions is empty then validation passes", func(t *testing.T) {
		config := newValidPickupConfig()
		config.Instructions = "" // Instructions is optional

		err := config.Validate()

		assert.NoError(t, err)
	})
}
