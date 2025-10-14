package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ProductService interface {
	Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error)
}
