package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CursorData represents the internal structure of a pagination cursor.
// This structure is encoded as base64 and sent to clients as an opaque string.
type CursorData struct {
	ID        int         `json:"id"`          // Required: unique identifier for tie-breaking
	SortValue interface{} `json:"v,omitempty"` // Optional: value of the field being sorted by
}

// EncodeCursor converts cursor data to an opaque base64 string.
// This is what gets sent to clients in API responses.
func EncodeCursor(data CursorData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor data: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(jsonData)
	return encoded, nil
}

// DecodeCursor converts an opaque cursor string back to cursor data.
// Returns nil, nil for empty cursor (first page).
// Returns nil, error for invalid cursor format.
func DecodeCursor(cursor string) (*CursorData, error) {
	if cursor == "" {
		return nil, nil
	}

	jsonData, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	var data CursorData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("invalid cursor data: %w", err)
	}

	if data.ID <= 0 {
		return nil, fmt.Errorf("invalid cursor: missing or invalid id")
	}

	return &data, nil
}
