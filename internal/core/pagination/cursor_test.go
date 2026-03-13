package pagination

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// EncodeCursor
// =============================================================================

func TestEncodeCursor(t *testing.T) {
	t.Run("when valid data then returns non-empty base64 string", func(t *testing.T) {
		data := CursorData{ID: 42}

		encoded, err := EncodeCursor(data)

		require.NoError(t, err)
		assert.NotEmpty(t, encoded)

		// Verify it's valid base64
		_, decErr := base64.URLEncoding.DecodeString(encoded)
		assert.NoError(t, decErr)
	})

	t.Run("when data has SortValue then includes it in encoding", func(t *testing.T) {
		data := CursorData{ID: 10, SortValue: "my_code"}

		encoded, err := EncodeCursor(data)

		require.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})

	t.Run("when SortValue is nil then encodes without it", func(t *testing.T) {
		data := CursorData{ID: 5}

		encoded, err := EncodeCursor(data)

		require.NoError(t, err)

		// Decode and verify SortValue is absent
		decoded, err := DecodeCursor(encoded)
		require.NoError(t, err)
		assert.Equal(t, 5, decoded.ID)
		assert.Nil(t, decoded.SortValue)
	})
}

// =============================================================================
// DecodeCursor
// =============================================================================

func TestDecodeCursor(t *testing.T) {
	t.Run("when empty string then returns nil nil", func(t *testing.T) {
		result, err := DecodeCursor("")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("when valid encoded cursor then returns correct data", func(t *testing.T) {
		original := CursorData{ID: 42, SortValue: "test_value"}
		encoded, _ := EncodeCursor(original)

		result, err := DecodeCursor(encoded)

		require.NoError(t, err)
		assert.Equal(t, 42, result.ID)
		assert.Equal(t, "test_value", result.SortValue)
	})

	t.Run("when invalid base64 then returns error", func(t *testing.T) {
		result, err := DecodeCursor("!!!not-base64!!!")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid cursor format")
	})

	t.Run("when valid base64 but invalid JSON then returns error", func(t *testing.T) {
		encoded := base64.URLEncoding.EncodeToString([]byte("not json"))

		result, err := DecodeCursor(encoded)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid cursor data")
	})

	t.Run("when ID is zero then returns error", func(t *testing.T) {
		encoded := base64.URLEncoding.EncodeToString([]byte(`{"id":0}`))

		result, err := DecodeCursor(encoded)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid cursor")
	})

	t.Run("when ID is negative then returns error", func(t *testing.T) {
		encoded := base64.URLEncoding.EncodeToString([]byte(`{"id":-5}`))

		result, err := DecodeCursor(encoded)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid cursor")
	})
}

// =============================================================================
// Roundtrip
// =============================================================================

func TestEncodeDecode_Roundtrip(t *testing.T) {
	t.Run("when encode then decode then returns original data", func(t *testing.T) {
		original := CursorData{ID: 100, SortValue: "2024-01-15T10:30:00Z"}

		encoded, err := EncodeCursor(original)
		require.NoError(t, err)

		decoded, err := DecodeCursor(encoded)
		require.NoError(t, err)

		assert.Equal(t, original.ID, decoded.ID)
		assert.Equal(t, original.SortValue, decoded.SortValue)
	})
}
