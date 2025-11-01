package services

import "github.com/mlgaray/ecommerce_api/internal/core/ports"

type PaginationService[T ports.Identifiable] struct{}

func NewPaginationService[T ports.Identifiable]() *PaginationService[T] {
	return &PaginationService[T]{}
}

// BuildCursorPagination builds cursor-based pagination metadata
// Usage examples:
//   - For products: nextCursor, hasMore := paginationService.BuildCursorPagination(products, limit)
//   - For categories: nextCursor, hasMore := paginationService.BuildCursorPagination(categories, limit)
func (p *PaginationService[T]) BuildCursorPagination(
	items []T,
	limit int,
) (nextCursor int, hasMore bool) {
	if len(items) > 0 && len(items) == limit {
		// Get the ID of the last item as the next cursor
		lastItem := items[len(items)-1]
		nextCursor = lastItem.GetID()
		hasMore = true
	}
	return nextCursor, hasMore
}
