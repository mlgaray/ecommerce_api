package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidDeliveryZone() *DeliveryZone {
	return &DeliveryZone{
		Name: "Centro",
		Cost: 500.00,
	}
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestDeliveryZone_Validate(t *testing.T) {
	t.Run("when zone is valid then returns nil", func(t *testing.T) {
		zone := newValidDeliveryZone()

		err := zone.Validate()

		assert.NoError(t, err)
	})

	t.Run("when name is empty then returns error", func(t *testing.T) {
		zone := newValidDeliveryZone()
		zone.Name = ""

		err := zone.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.DeliveryZoneNameIsRequired, validationErr.Message)
	})

	t.Run("when cost is negative then returns error", func(t *testing.T) {
		zone := newValidDeliveryZone()
		zone.Cost = -100.00

		err := zone.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.DeliveryZoneCostCannotBeNegative, validationErr.Message)
	})

	t.Run("when cost is zero then validation passes", func(t *testing.T) {
		zone := newValidDeliveryZone()
		zone.Cost = 0 // Free delivery is allowed

		err := zone.Validate()

		assert.NoError(t, err)
	})

	t.Run("when cost is positive then validation passes", func(t *testing.T) {
		zone := newValidDeliveryZone()
		zone.Cost = 1000.50

		err := zone.Validate()

		assert.NoError(t, err)
	})
}
