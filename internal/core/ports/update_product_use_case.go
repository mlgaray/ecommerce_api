package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type UpdateProductUseCase interface {
	Execute(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte) error
}
