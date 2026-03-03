package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// parseTimezone
// =============================================================================

func TestParseTimezone(t *testing.T) {
	t.Run("when nil then returns UTC", func(t *testing.T) {
		// Act
		loc, name := parseTimezone(nil)

		// Assert
		assert.Equal(t, time.UTC, loc)
		assert.Equal(t, "UTC", name)
	})

	t.Run("when empty string then returns UTC", func(t *testing.T) {
		// Arrange
		tz := ""

		// Act
		loc, name := parseTimezone(&tz)

		// Assert
		assert.Equal(t, time.UTC, loc)
		assert.Equal(t, "UTC", name)
	})

	t.Run("when valid timezone then returns correct location", func(t *testing.T) {
		// Arrange
		tz := "America/New_York"

		// Act
		loc, name := parseTimezone(&tz)

		// Assert
		expectedLoc, _ := time.LoadLocation("America/New_York")
		assert.Equal(t, expectedLoc, loc)
		assert.Equal(t, "America/New_York", name)
	})

	t.Run("when invalid timezone then falls back to UTC", func(t *testing.T) {
		// Arrange
		tz := "Invalid/Zone"

		// Act
		loc, name := parseTimezone(&tz)

		// Assert
		assert.Equal(t, time.UTC, loc)
		assert.Equal(t, "UTC", name)
	})
}
