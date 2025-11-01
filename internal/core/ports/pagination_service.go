package ports

// Identifiable represents any entity that has an ID field
// Used for generic operations like pagination that require entity identification
type Identifiable interface {
	GetID() int
}

// PaginationService provides reusable cursor-based pagination logic
// Generic interface that works with any type implementing Identifiable
type PaginationService[T Identifiable] interface {
	BuildCursorPagination(items []T, limit int) (nextCursor int, hasMore bool)
}
