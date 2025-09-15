package ports

import "context"

type AuthService interface {
	// HashPassword(ctx context.Context, password string) (string, error)
	ComparePassword(ctx context.Context, hashedPassword, password string) error
}
