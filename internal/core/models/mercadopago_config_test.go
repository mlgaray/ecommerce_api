package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidMercadoPagoConfig() *MercadoPagoConfig {
	return &MercadoPagoConfig{
		AccessToken: "TEST-1234567890123456-123456-abcdefghijklmnopqrstuvwxyz123456-123456789",
		PublicKey:   "TEST-12345678-1234-1234-1234-123456789012",
		UserID:      "123456789",
	}
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestMercadoPagoConfig_Validate(t *testing.T) {
	t.Run("when config is valid then returns nil", func(t *testing.T) {
		config := newValidMercadoPagoConfig()

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("when access_token is empty then returns error", func(t *testing.T) {
		config := newValidMercadoPagoConfig()
		config.AccessToken = ""

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.MercadoPagoAccessTokenIsRequired, validationErr.Message)
	})

	t.Run("when public_key is empty then returns error", func(t *testing.T) {
		config := newValidMercadoPagoConfig()
		config.PublicKey = ""

		err := config.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.MercadoPagoPublicKeyIsRequired, validationErr.Message)
	})

	t.Run("when user_id is empty then validation passes", func(t *testing.T) {
		config := newValidMercadoPagoConfig()
		config.UserID = "" // UserID is optional

		err := config.Validate()

		assert.NoError(t, err)
	})
}
