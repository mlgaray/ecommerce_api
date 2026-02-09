package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// OptionsEqualForOrder Tests
// =============================================================================

func TestVariant_OptionsEqualForOrder(t *testing.T) {
	t.Run("when no frontend options then returns true", func(t *testing.T) {
		variant := &Variant{
			ID:   1,
			Name: "Tamaño",
			Options: []*Option{
				{ID: 1, Name: "Chico", Price: 0},
				{ID: 2, Name: "Grande", Price: 1500},
			},
		}

		assert.True(t, variant.OptionsEqualForOrder([]*Option{}))
	})

	t.Run("when all options match then returns true", func(t *testing.T) {
		variant := &Variant{
			ID:   1,
			Name: "Tamaño",
			Options: []*Option{
				{ID: 1, Name: "Chico", Price: 0},
				{ID: 2, Name: "Grande", Price: 1500},
				{ID: 3, Name: "Extra Grande", Price: 2000},
			},
		}
		frontendOptions := []*Option{
			{ID: 2, Name: "Grande", Price: 1500},
		}

		assert.True(t, variant.OptionsEqualForOrder(frontendOptions))
	})

	t.Run("when multiple options match then returns true", func(t *testing.T) {
		variant := &Variant{
			ID:   1,
			Name: "Extras",
			Options: []*Option{
				{ID: 1, Name: "Extra bacon", Price: 800},
				{ID: 2, Name: "Extra queso", Price: 800},
				{ID: 3, Name: "Extra lechuga", Price: 500},
			},
		}
		frontendOptions := []*Option{
			{ID: 1, Name: "Extra bacon", Price: 800},
			{ID: 2, Name: "Extra queso", Price: 800},
		}

		assert.True(t, variant.OptionsEqualForOrder(frontendOptions))
	})

	t.Run("when option not found in variant then returns false", func(t *testing.T) {
		variant := &Variant{
			ID:   1,
			Name: "Tamaño",
			Options: []*Option{
				{ID: 1, Name: "Chico", Price: 0},
			},
		}
		frontendOptions := []*Option{
			{ID: 999, Name: "Unknown", Price: 0},
		}

		assert.False(t, variant.OptionsEqualForOrder(frontendOptions))
	})

	t.Run("when option name differs then returns false", func(t *testing.T) {
		variant := &Variant{
			ID:   1,
			Name: "Tamaño",
			Options: []*Option{
				{ID: 1, Name: "Chico", Price: 0},
			},
		}
		frontendOptions := []*Option{
			{ID: 1, Name: "Wrong Name", Price: 0},
		}

		assert.False(t, variant.OptionsEqualForOrder(frontendOptions))
	})

	t.Run("when option price differs then returns false", func(t *testing.T) {
		variant := &Variant{
			ID:   1,
			Name: "Tamaño",
			Options: []*Option{
				{ID: 1, Name: "Chico", Price: 0},
			},
		}
		frontendOptions := []*Option{
			{ID: 1, Name: "Chico", Price: 9999},
		}

		assert.False(t, variant.OptionsEqualForOrder(frontendOptions))
	})

	t.Run("when variant has no options then frontend option not found returns false", func(t *testing.T) {
		variant := &Variant{
			ID:      1,
			Name:    "Tamaño",
			Options: []*Option{},
		}
		frontendOptions := []*Option{
			{ID: 1, Name: "Grande", Price: 1500},
		}

		assert.False(t, variant.OptionsEqualForOrder(frontendOptions))
	})

	t.Run("when variant has nil options then frontend option not found returns false", func(t *testing.T) {
		variant := &Variant{
			ID:      1,
			Name:    "Tamaño",
			Options: nil,
		}
		frontendOptions := []*Option{
			{ID: 1, Name: "Grande", Price: 1500},
		}

		assert.False(t, variant.OptionsEqualForOrder(frontendOptions))
	})

	t.Run("when nil frontend options then returns true", func(t *testing.T) {
		variant := &Variant{
			ID:   1,
			Name: "Tamaño",
			Options: []*Option{
				{ID: 1, Name: "Grande", Price: 1500},
			},
		}

		assert.True(t, variant.OptionsEqualForOrder(nil))
	})
}
