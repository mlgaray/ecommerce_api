package ports

import "context"

// StoreVisitRepository handles persistence for store visit tracking.
type StoreVisitRepository interface {
	// RecordVisit inserts a new visit record for the given shop.
	RecordVisit(ctx context.Context, shopID int) error
}
