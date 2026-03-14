package ports

import "context"

// CheckSlugAvailabilityUseCase checks if a store slug is available for registration.
type CheckSlugAvailabilityUseCase interface {
	Execute(ctx context.Context, slug string) (bool, error)
}
