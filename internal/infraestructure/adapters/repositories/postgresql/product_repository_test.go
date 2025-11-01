package postgresql

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestMain(m *testing.M) {
	// Initialize logger before running tests
	logs.Init()

	// Run tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}

func TestNewProductRepository(t *testing.T) {
	t.Run("when called then returns ProductRepository", func(t *testing.T) {
		// Arrange
		db, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mockDbConnection := mocks.NewDataBaseConnection(t)
		mockDbConnection.EXPECT().Connect().Return(db)

		// Act
		repo := NewProductRepository(mockDbConnection)

		// Assert
		assert.NotNil(t, repo)
		assert.IsType(t, &ProductRepository{}, repo)
	})
}

func TestProductRepository_Create(t *testing.T) {
	t.Run("when product is created successfully without transaction then returns product with ID", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		product := &models.Product{
			Name:             "Test Product",
			Description:      "Test Description",
			Price:            99.99,
			Stock:            10,
			IsActive:         true,
			IsHighlighted:    false,
			IsPromotional:    false,
			PromotionalPrice: 0,
			Category: &models.Category{
				ID: 1,
			},
			Images:   []string{"http://example.com/image1.jpg"},
			Variants: []*models.Variant{},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO products`).
			WithArgs(
				product.Name,
				product.Description,
				product.Price,
				product.Stock,
				product.IsActive,
				product.IsHighlighted,
				product.IsPromotional,
				product.PromotionalPrice,
				product.Category.ID,
				shopID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectExec(`INSERT INTO product_images`).
			WithArgs("http://example.com/image1.jpg", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		repo := &ProductRepository{db: db}

		// Act
		createdProduct, err := repo.Create(ctx, product, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, createdProduct)
		assert.Equal(t, 1, createdProduct.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when product is created successfully with transaction then returns product with ID", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.WithValue(context.Background(), TxContextKey, tx)
		shopID := 1
		product := &models.Product{
			Name:             "Test Product",
			Description:      "Test Description",
			Price:            99.99,
			Stock:            10,
			IsActive:         true,
			IsHighlighted:    false,
			IsPromotional:    false,
			PromotionalPrice: 0,
			Category: &models.Category{
				ID: 1,
			},
			Images:   []string{},
			Variants: []*models.Variant{},
		}

		mock.ExpectQuery(`INSERT INTO products`).
			WithArgs(
				product.Name,
				product.Description,
				product.Price,
				product.Stock,
				product.IsActive,
				product.IsHighlighted,
				product.IsPromotional,
				product.PromotionalPrice,
				product.Category.ID,
				shopID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		repo := &ProductRepository{db: db}

		// Act
		createdProduct, err := repo.Create(ctx, product, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, createdProduct)
		assert.Equal(t, 1, createdProduct.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when product creation fails at begin transaction then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		product := &models.Product{
			Name: "Test Product",
		}

		expectedError := errors.New("begin transaction failed")
		mock.ExpectBegin().WillReturnError(expectedError)

		repo := &ProductRepository{db: db}

		// Act
		createdProduct, err := repo.Create(ctx, product, 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, createdProduct)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when product insertion fails then returns error and rolls back", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			Category: &models.Category{
				ID: 1,
			},
		}

		expectedError := errors.New("insert product failed")
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO products`).
			WithArgs(
				product.Name,
				product.Description,
				product.Price,
				product.Stock,
				product.IsActive,
				product.IsHighlighted,
				product.IsPromotional,
				product.PromotionalPrice,
				product.Category.ID,
				shopID,
			).
			WillReturnError(expectedError)
		mock.ExpectRollback()

		repo := &ProductRepository{db: db}

		// Act
		createdProduct, err := repo.Create(ctx, product, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, createdProduct)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when commit fails then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			Category: &models.Category{
				ID: 1,
			},
			Images:   []string{},
			Variants: []*models.Variant{},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO products`).
			WithArgs(
				product.Name,
				product.Description,
				product.Price,
				product.Stock,
				product.IsActive,
				product.IsHighlighted,
				product.IsPromotional,
				product.PromotionalPrice,
				product.Category.ID,
				shopID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		expectedError := errors.New("commit failed")
		mock.ExpectCommit().WillReturnError(expectedError)
		// Note: No ExpectRollback here because when Commit fails,
		// the deferred Rollback is a no-op (transaction is already closed)

		repo := &ProductRepository{db: db}

		// Act
		createdProduct, err := repo.Create(ctx, product, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, createdProduct)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductRepository_InsertProductImages(t *testing.T) {
	t.Run("when images are inserted successfully then returns no error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.Background()
		productID := 1
		imageURLs := []string{
			"http://example.com/image1.jpg",
			"http://example.com/image2.jpg",
		}

		for _, imageURL := range imageURLs {
			mock.ExpectExec(`INSERT INTO product_images`).
				WithArgs(imageURL, productID).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		repo := &ProductRepository{db: db}

		// Act
		err = repo.insertProductImages(ctx, tx, productID, imageURLs)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when image insertion fails then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.Background()
		productID := 1
		imageURLs := []string{"http://example.com/image1.jpg"}

		expectedError := errors.New("insert image failed")
		mock.ExpectExec(`INSERT INTO product_images`).
			WithArgs(imageURLs[0], productID).
			WillReturnError(expectedError)

		repo := &ProductRepository{db: db}

		// Act
		err = repo.insertProductImages(ctx, tx, productID, imageURLs)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductRepository_InsertProductVariants(t *testing.T) {
	t.Run("when variants are inserted successfully then returns no error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.Background()
		productID := 1
		variants := []*models.Variant{
			{
				Name:          "Size",
				Order:         1,
				SelectionType: "single",
				MaxSelections: 1,
				Options: []*models.Option{
					{
						Name:  "Small",
						Price: 0.0,
						Order: 1,
					},
					{
						Name:  "Large",
						Price: 5.0,
						Order: 2,
					},
				},
			},
		}

		// Expect variant insertion
		mock.ExpectQuery(`INSERT INTO product_variants`).
			WithArgs(
				variants[0].Name,
				variants[0].Order,
				variants[0].SelectionType,
				variants[0].MaxSelections,
				productID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect option insertions
		for _, option := range variants[0].Options {
			mock.ExpectExec(`INSERT INTO variant_options`).
				WithArgs(
					option.Name,
					option.Price,
					option.Order,
					1, // variant ID
				).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		repo := &ProductRepository{db: db}

		// Act
		err = repo.insertProductVariants(ctx, tx, productID, variants)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, variants[0].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when variant insertion fails then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.Background()
		productID := 1
		variants := []*models.Variant{
			{
				Name:          "Size",
				Order:         1,
				SelectionType: "single",
				MaxSelections: 1,
				Options:       []*models.Option{},
			},
		}

		expectedError := errors.New("insert variant failed")
		mock.ExpectQuery(`INSERT INTO product_variants`).
			WithArgs(
				variants[0].Name,
				variants[0].Order,
				variants[0].SelectionType,
				variants[0].MaxSelections,
				productID,
			).
			WillReturnError(expectedError)

		repo := &ProductRepository{db: db}

		// Act
		err = repo.insertProductVariants(ctx, tx, productID, variants)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when variant option insertion fails then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.Background()
		productID := 1
		variants := []*models.Variant{
			{
				Name:          "Size",
				Order:         1,
				SelectionType: "single",
				MaxSelections: 1,
				Options: []*models.Option{
					{
						Name:  "Small",
						Price: 0.0,
						Order: 1,
					},
				},
			},
		}

		// Expect variant insertion
		mock.ExpectQuery(`INSERT INTO product_variants`).
			WithArgs(
				variants[0].Name,
				variants[0].Order,
				variants[0].SelectionType,
				variants[0].MaxSelections,
				productID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect option insertion to fail
		expectedError := errors.New("insert option failed")
		mock.ExpectExec(`INSERT INTO variant_options`).
			WithArgs(
				variants[0].Options[0].Name,
				variants[0].Options[0].Price,
				variants[0].Options[0].Order,
				1, // variant ID
			).
			WillReturnError(expectedError)

		repo := &ProductRepository{db: db}

		// Act
		err = repo.insertProductVariants(ctx, tx, productID, variants)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductRepository_CreateComplexProduct(t *testing.T) {
	t.Run("when creating complex product with images and variants then all data is inserted correctly", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		product := &models.Product{
			Name:             "Complex Product",
			Description:      "Product with images and variants",
			Price:            99.99,
			Stock:            50,
			IsActive:         true,
			IsHighlighted:    true,
			IsPromotional:    true,
			PromotionalPrice: 79.99,
			Category: &models.Category{
				ID: 2,
			},
			Images: []string{
				"http://example.com/image1.jpg",
				"http://example.com/image2.jpg",
			},
			Variants: []*models.Variant{
				{
					Name:          "Size",
					Order:         1,
					SelectionType: "single",
					MaxSelections: 1,
					Options: []*models.Option{
						{Name: "Small", Price: 0.0, Order: 1},
						{Name: "Medium", Price: 2.0, Order: 2},
						{Name: "Large", Price: 5.0, Order: 3},
					},
				},
				{
					Name:          "Color",
					Order:         2,
					SelectionType: "single",
					MaxSelections: 1,
					Options: []*models.Option{
						{Name: "Red", Price: 0.0, Order: 1},
						{Name: "Blue", Price: 0.0, Order: 2},
					},
				},
			},
		}

		mock.ExpectBegin()

		// Expect product insertion
		mock.ExpectQuery(`INSERT INTO products`).
			WithArgs(
				product.Name,
				product.Description,
				product.Price,
				product.Stock,
				product.IsActive,
				product.IsHighlighted,
				product.IsPromotional,
				product.PromotionalPrice,
				product.Category.ID,
				shopID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect image insertions
		for _, imageURL := range product.Images {
			mock.ExpectExec(`INSERT INTO product_images`).
				WithArgs(imageURL, 1).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		// Expect variant and option insertions
		variantID := 1
		for _, variant := range product.Variants {
			mock.ExpectQuery(`INSERT INTO product_variants`).
				WithArgs(
					variant.Name,
					variant.Order,
					variant.SelectionType,
					variant.MaxSelections,
					1, // product ID
				).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(variantID))

			for _, option := range variant.Options {
				mock.ExpectExec(`INSERT INTO variant_options`).
					WithArgs(
						option.Name,
						option.Price,
						option.Order,
						variantID,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}
			variantID++
		}

		mock.ExpectCommit()

		repo := &ProductRepository{db: db}

		// Act
		createdProduct, err := repo.Create(ctx, product, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, createdProduct)
		assert.Equal(t, 1, createdProduct.ID)
		assert.Equal(t, 1, createdProduct.Variants[0].ID)
		assert.Equal(t, 2, createdProduct.Variants[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductRepository_GetAllByShopID(t *testing.T) {
	t.Run("when getting products without cursor then returns first page", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 20
		cursor := 0

		imagesJSON := `["http://example.com/image1.jpg","http://example.com/image2.jpg"]`
		variantsJSON := `[{"id":1,"name":"Size","order":1,"selection_type":"single","max_selections":1,"options":[{"id":1,"name":"Small","price":0,"order":1}]}]`

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		}).
			AddRow(
				1, "Product 1", "Description 1", 99.99, 10, 5,
				true, false, false, 0.0,
				1, "Category 1", "Category Description",
				[]byte(imagesJSON), []byte(variantsJSON),
			).
			AddRow(
				2, "Product 2", "Description 2", 149.99, 20, 10,
				true, true, true, 129.99,
				2, "Category 2", "Category Description 2",
				[]byte("[]"), []byte("[]"),
			)

		mock.ExpectQuery(`SELECT(.+)FROM products p(.+)WHERE p.shop_id = \$1 AND p.is_active = true(.+)ORDER BY p.id DESC(.+)LIMIT \$2`).
			WithArgs(shopID, limit).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Len(t, products, 2)
		assert.Equal(t, 1, products[0].ID)
		assert.Equal(t, "Product 1", products[0].Name)
		assert.Len(t, products[0].Images, 2)
		assert.Len(t, products[0].Variants, 1)
		assert.Equal(t, 2, products[1].ID)
		assert.Equal(t, "Product 2", products[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when getting products with cursor then returns paginated results", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 20
		cursor := 100

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		}).
			AddRow(
				99, "Product 99", "Description 99", 79.99, 15, 5,
				true, false, false, 0.0,
				1, "Category 1", "",
				[]byte("[]"), []byte("[]"),
			)

		mock.ExpectQuery(`SELECT(.+)FROM products p(.+)WHERE p.shop_id = \$1 AND p.is_active = true AND p.id < \$2(.+)ORDER BY p.id DESC(.+)LIMIT \$3`).
			WithArgs(shopID, cursor, limit).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Len(t, products, 1)
		assert.Equal(t, 99, products[0].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when limit is zero then uses default limit of 20", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 0
		cursor := 0

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		})

		// Expect default limit of 20
		mock.ExpectQuery(`SELECT(.+)FROM products p(.+)WHERE p.shop_id = \$1 AND p.is_active = true(.+)ORDER BY p.id DESC(.+)LIMIT \$2`).
			WithArgs(shopID, 20).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Len(t, products, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when limit exceeds 100 then uses max limit of 100", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 200
		cursor := 0

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		})

		// Expect max limit of 100
		mock.ExpectQuery(`SELECT(.+)FROM products p(.+)WHERE p.shop_id = \$1 AND p.is_active = true(.+)ORDER BY p.id DESC(.+)LIMIT \$2`).
			WithArgs(shopID, 100).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when no products found then returns empty slice", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 999
		limit := 20
		cursor := 0

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		})

		mock.ExpectQuery(`SELECT(.+)FROM products p(.+)WHERE p.shop_id = \$1 AND p.is_active = true(.+)ORDER BY p.id DESC(.+)LIMIT \$2`).
			WithArgs(shopID, limit).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Len(t, products, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when query fails then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 20
		cursor := 0

		expectedError := errors.New("database query failed")
		mock.ExpectQuery(`SELECT(.+)FROM products p`).
			WithArgs(shopID, limit).
			WillReturnError(expectedError)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, products)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when scan fails then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 20
		cursor := 0

		// Return wrong number of columns to cause scan error
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Product 1")

		mock.ExpectQuery(`SELECT(.+)FROM products p`).
			WithArgs(shopID, limit).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, products)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when images JSON is invalid then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 20
		cursor := 0

		invalidImagesJSON := `[invalid json`

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		}).
			AddRow(
				1, "Product 1", "Description 1", 99.99, 10, 5,
				true, false, false, 0.0,
				1, "Category 1", "",
				[]byte(invalidImagesJSON), []byte("[]"),
			)

		mock.ExpectQuery(`SELECT(.+)FROM products p`).
			WithArgs(shopID, limit).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, products)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when variants JSON is invalid then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 20
		cursor := 0

		invalidVariantsJSON := `[invalid json`

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		}).
			AddRow(
				1, "Product 1", "Description 1", 99.99, 10, 5,
				true, false, false, 0.0,
				1, "Category 1", "",
				[]byte("[]"), []byte(invalidVariantsJSON),
			)

		mock.ExpectQuery(`SELECT(.+)FROM products p`).
			WithArgs(shopID, limit).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, products)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when rows iteration error occurs then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		limit := 20
		cursor := 0

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		}).
			AddRow(
				1, "Product 1", "Description 1", 99.99, 10, 5,
				true, false, false, 0.0,
				1, "Category 1", "",
				[]byte("[]"), []byte("[]"),
			).
			RowError(0, errors.New("rows iteration error"))

		mock.ExpectQuery(`SELECT(.+)FROM products p`).
			WithArgs(shopID, limit).
			WillReturnRows(rows)

		repo := &ProductRepository{db: db}

		// Act
		products, err := repo.GetAllByShopID(ctx, shopID, limit, cursor)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, products)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
