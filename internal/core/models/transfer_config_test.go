package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidTransferConfig() *TransferConfig {
	return &TransferConfig{
		CBU:       "0123456789012345678901", // 22 characters
		CUIL:      "20-12345678-9",
		Alias:     "MI.ALIAS.MP",
		OwnerName: "John Doe",
	}
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestTransferConfig_Validate(t *testing.T) {
	t.Run("when config is valid then returns nil", func(t *testing.T) {
		config := newValidTransferConfig()

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("when CBU is empty then returns error", func(t *testing.T) {
		config := newValidTransferConfig()
		config.CBU = ""

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.TransferCBUIsRequired, validationErr.Message)
	})

	t.Run("when CBU has less than 22 characters then returns error", func(t *testing.T) {
		config := newValidTransferConfig()
		config.CBU = "012345678901234567890" // 21 characters

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.TransferCBUInvalidLength, validationErr.Message)
	})

	t.Run("when CBU has more than 22 characters then returns error", func(t *testing.T) {
		config := newValidTransferConfig()
		config.CBU = "01234567890123456789012" // 23 characters

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.TransferCBUInvalidLength, validationErr.Message)
	})

	t.Run("when CUIL is empty then returns error", func(t *testing.T) {
		config := newValidTransferConfig()
		config.CUIL = ""

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.TransferCUILIsRequired, validationErr.Message)
	})

	t.Run("when owner_name is empty then returns error", func(t *testing.T) {
		config := newValidTransferConfig()
		config.OwnerName = ""

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.TransferOwnerNameIsRequired, validationErr.Message)
	})

	t.Run("when alias is empty then validation passes", func(t *testing.T) {
		config := newValidTransferConfig()
		config.Alias = "" // Alias is optional

		err := config.Validate()

		assert.NoError(t, err)
	})
}
