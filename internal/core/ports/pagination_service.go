package ports

// Identifiable represents any entity that has an ID field
// Used for generic operations like pagination that require entity identification
type Identifiable interface {
	GetID() int
}

// PaginationService provides reusable cursor-based pagination logic
// Generic interface that works with any type implementing both Identifiable and Sortable
type PaginationService[T interface {
	Identifiable
	Sortable
}] interface {
	// ApplyPagination applies the complete pagination strategy:
	// 1. Detects if there are more pages (hasMore)
	// 2. Builds the cursor for the next page (if hasMore is true)
	// 3. Trims the extra item from LIMIT+1 strategy (if present)
	//
	// Parameters:
	//   - items: List of items fetched with LIMIT+1 strategy
	//   - limit: Original limit requested (without +1)
	//   - sortBy: Field name used for sorting (e.g., "price", "name", "created_at")
	//
	// Returns:
	//   - trimmedItems: List with max 'limit' items (extra item removed if present)
	//   - nextCursor: Opaque base64-encoded cursor for next page (empty if no more pages)
	//   - hasMore: true if there are more pages available
	ApplyPagination(items []T, limit int, sortBy string) (trimmedItems []T, nextCursor string, hasMore bool)
}
