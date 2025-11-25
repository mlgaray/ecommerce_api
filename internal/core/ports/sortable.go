package ports

// Sortable defines the contract for entities that can be sorted
// Models implementing this interface can be used with PaginationService
type Sortable interface {
	// GetSortValue returns the value of the specified sort field
	// Returns nil if the field doesn't exist or isn't supported for sorting
	//
	// Example:
	//   value := product.GetSortValue("price")  // Returns product.Price
	//   value := product.GetSortValue("name")   // Returns product.Name
	GetSortValue(sortBy string) interface{}
}
