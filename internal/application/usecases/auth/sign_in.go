package auth

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignInUseCase struct {
	userService    ports.UserService
	tokenService   ports.TokenService
	shopRepository ports.ShopRepository
}

func NewSignInUseCase(userService ports.UserService, tokenService ports.TokenService, shopRepository ports.ShopRepository) ports.SignInUseCase {
	return &SignInUseCase{
		userService:    userService,
		tokenService:   tokenService,
		shopRepository: shopRepository,
	}
}

func (uc *SignInUseCase) Execute(ctx context.Context, user *models.User) (string, error) {
	storedUser, err := uc.userService.GetByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	_, err = uc.userService.ValidateCredentials(ctx, user, storedUser.Password)
	if err != nil {
		return "", err
	}

	// Get user's shops to include in token (use storedUser.ID, not input user)
	shops, err := uc.shopRepository.GetShopsByUserID(ctx, storedUser.ID)
	if err != nil {
		return "", err
	}

	// Extract shop IDs for token payload
	shopIDs := make([]int, len(shops))
	for i, shop := range shops {
		shopIDs[i] = shop.ID
	}

	token, err := uc.tokenService.Generate(ctx, storedUser, shopIDs)
	if err != nil {
		return "", err
	}

	return token, nil
}
