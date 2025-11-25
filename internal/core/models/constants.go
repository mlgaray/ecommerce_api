package models

// Sort field constants for product queries
const (
	SortByPrice     = "price"
	SortByName      = "name"
	SortByCreatedAt = "created_at"
	SortByID        = "id"
)

// Sort order constants
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// Pagination defaults
const (
	DefaultLimit = 20
	MaxLimit     = 100
)
