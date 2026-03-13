package services

import (
	"github.com/mlgaray/ecommerce_api/internal/core/pagination"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// PaginationService handles cursor-based pagination logic
// Type parameter T must implement both Identifiable and Sortable interfaces
type PaginationService[T interface {
	ports.Identifiable
	ports.Sortable
}] struct{}

func NewPaginationService[T interface {
	ports.Identifiable
	ports.Sortable
}]() *PaginationService[T] {
	return &PaginationService[T]{}
}

// ApplyPagination applies complete cursor-based pagination to a list of items
// Encapsulates the entire LIMIT+1 strategy:
//  1. Detects if there are more pages
//  2. Builds the next cursor from the last valid item
//  3. Trims the extra item (if it exists)
//
// Parameters:
//   - items: Fetched items from repository (should contain limit + 1 if there are more pages)
//   - limit: The requested page size
//   - sortBy: The field being sorted by ("id", "created_at", "price", "name", etc.)
//
// Returns:
//   - trimmedItems: Items with the extra item removed (maximum of 'limit' items)
//   - nextCursor: Opaque base64-encoded cursor for the next page (empty string if no more pages)
//   - hasMore: true if there are more pages available
//
// Example Usage:
//
//	// Repository returned 21 products (limit was 20, so there's 1 extra)
//	products, cursor, hasMore := service.ApplyPagination(products, 20, "price")
//	// Returns: 20 products, cursor="...", hasMore=true
//
//	// Repository returned 15 products (limit was 20, no extras)
//	products, cursor, hasMore := service.ApplyPagination(products, 20, "created_at")
//	// Returns: 15 products, cursor="", hasMore=false
func (p *PaginationService[T]) ApplyPagination(
	items []T,
	limit int,
	sortBy string,
) (trimmedItems []T, nextCursor string, hasMore bool) {
	// Empty list - no pagination needed
	if len(items) == 0 {
		return items, "", false
	}

	// 1. Detect if there are more pages (LIMIT+1 strategy)
	hasMore = p.hasMorePages(items, limit)

	// 2. Get the last valid item (the one that will be shown to the user)
	lastValidItem := p.getLastValidItem(items, limit, hasMore)

	// 3. Build cursor only if there are more pages
	nextCursor = ""
	if hasMore {
		nextCursor = p.buildCursor(lastValidItem, sortBy)
	}

	// 4. Trim the extra item if present
	trimmedItems = p.trimItems(items, limit, hasMore)

	return trimmedItems, nextCursor, hasMore
}

// ===== PRIVATE HELPER METHODS (Each with Single Responsibility) =====

// hasMorePages detects if there are more pages using the LIMIT+1 strategy
// If we have more items than the requested limit, there are more pages
func (p *PaginationService[T]) hasMorePages(items []T, limit int) bool {
	return len(items) > limit
}

// getLastValidItem returns the last item that will be shown to the user
// This is used to build the cursor for the next page
//
// Logic:
//   - If hasMore=true: return item at position (limit-1) - the last item before the extra
//   - If hasMore=false: return the actual last item in the array
//
// Precondition: items must not be empty (caller should check this)
func (p *PaginationService[T]) getLastValidItem(items []T, limit int, hasMore bool) T {
	if hasMore && limit > 0 {
		// Last item before the extra (position limit-1)
		return items[limit-1]
	}
	// Actual last item (or fallback when limit <= 0)
	return items[len(items)-1]
}

// buildCursor constructs an opaque cursor from an item
// Uses the item's ID and sort field value
func (p *PaginationService[T]) buildCursor(item T, sortBy string) string {
	// Extract the sort value using the Sortable interface
	sortValue := item.GetSortValue(sortBy)

	// Build cursor data
	cursorData := pagination.CursorData{
		ID:        item.GetID(),
		SortValue: sortValue,
	}

	// Encode to base64
	encoded, err := pagination.EncodeCursor(cursorData)
	if err != nil {
		// Fallback to empty cursor on encoding error (should rarely happen)
		return ""
	}

	return encoded
}

// trimItems removes the extra item if it exists
// Returns a slice with at most 'limit' items
func (p *PaginationService[T]) trimItems(items []T, limit int, hasMore bool) []T {
	if hasMore {
		// Remove the extra item (LIMIT+1 -> LIMIT)
		return items[:limit]
	}
	// No extra item, return as-is
	return items
}
