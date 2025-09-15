package postgresql

import (
	"context"
	"database/sql"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type ShopSQLRepository struct {
	db *sql.DB
}

func (s *ShopSQLRepository) Create(ctx context.Context, shop *models.Shop) (*models.Shop, error) {
	// Extraer transacción del contexto si existe
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.createWithTx(ctx, tx, shop)
	}

	// Si no hay transacción, usar conexión directa
	return s.createWithDB(ctx, shop)
}

func (s *ShopSQLRepository) createWithTx(ctx context.Context, tx *sql.Tx, shop *models.Shop) (*models.Shop, error) {
	const query = `
		INSERT INTO shops (user_id, name, slug, email, phone, instagram, image)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var shopID int
	err := tx.QueryRowContext(ctx, query, shop.UserID, shop.Name, shop.Slug, shop.Email, shop.Phone, shop.Instagram, shop.Image).Scan(&shopID)
	if err != nil {
		return nil, err
	}

	shop.ID = shopID
	return shop, nil
}

func (s *ShopSQLRepository) createWithDB(ctx context.Context, shop *models.Shop) (*models.Shop, error) {
	const query = `
		INSERT INTO shops (user_id, name, slug, email, phone, instagram, image)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var shopID int
	err := s.db.QueryRowContext(ctx, query, shop.UserID, shop.Name, shop.Slug, shop.Email, shop.Phone, shop.Instagram, shop.Image).Scan(&shopID)
	if err != nil {
		return nil, err
	}

	shop.ID = shopID
	return shop, nil
}

//	func (s *ShopRepository) GetByID(ctx context.Context, shopID int) (*entities.Shop, error) {
//		query := `
//	       		SELECT
//	           		s.id, s.name, s.slug, s.email, s.phone, s.instagram, s.image,
//	           		a.id, a.text, a.place_id, a.ltd, a.lng
//	       		FROM shops s
//	       		JOIN addresses a ON s.id = a.shop_id
//	       		WHERE s.id = $1
//	   			`
//		row := s.DB.QueryRow(query, shopID)
//		shop := &entities.Shop{Address: &entities.Address{}}
//		err := row.Scan(
//			&shop.ID, &shop.Name, &shop.Slug, &shop.Email, &shop.Phone, &shop.Instagram, &shop.Image,
//			&shop.Address.ID, &shop.Address.Text, &shop.Address.PlaceID, &shop.Address.Ltd, &shop.Address.Lng,
//		)
//		if err != nil {
//			if err == sql.ErrNoRows {
//				//slog.Error("File: shop_repository.go | Func: GetByID() | SubFunc: row.Scan() | Msg: Shop with id %d no found | Err: %s", shopID, err.Error())
//				return nil, errors.New("Shop not found")
//			}
//			//slog.Error("File: shop_repository.go | Func: GetByID() | SubFunc: row.Scan() | Msg: Error scanning row | Err: %s", shopID, err.Error())
//			return nil, err
//		}
//		return shop, nil
//	}
//
//	func (s *ShopRepository) GetBySlug(ctx context.Context, slug string) (*entities.Shop, error) {
//		query := `
//			SELECT
//				s.id, s.name, s.slug, s.email, s.phone, s.instagram, s.image,
//				a.id, a.text, a.place_id, a.ltd, a.lng,
//				c.id, c.name, c.image
//			FROM shops s
//			JOIN addresses a ON s.id = a.shop_id
//			LEFT JOIN categories c ON s.id = c.shop_id
//			WHERE s.slug = $1
//		`
//
//		rows, err := s.DB.Query(query, slug)
//		if err != nil {
//			return nil, err
//		}
//		defer rows.Close()
//
//		shop := &entities.Shop{Address: &entities.Address{}, Categories: []*entities.Category{}}
//
//		for rows.Next() {
//			var category entities.Category
//
//			err := rows.Scan(
//				&shop.ID, &shop.Name, &shop.Slug, &shop.Email, &shop.Phone, &shop.Instagram, &shop.Image,
//				&shop.Address.ID, &shop.Address.Text, &shop.Address.PlaceID, &shop.Address.Ltd, &shop.Address.Lng,
//				&category.ID, &category.Name, &category.Image,
//			)
//			if err != nil {
//				return nil, err
//			}
//
//			// Agregar la categoría directamente a la lista
//			shop.Categories = append(shop.Categories, &category)
//		}
//
//		if err = rows.Err(); err != nil {
//			return nil, err
//		}
//
//		if shop.ID == 0 {
//			return nil, &e.NotFoundError{Message: "get_shop_by_slug_not_found"}
//		}
//
//		return shop, nil
//	}
//
//	func (s *ShopRepository) Update(ctx context.Context, shop *entities.Shop, shopID int) error {
//		var (
//			queryShop    = "UPDATE shops SET name = $1, email = $2, phone = $3, instagram = $4, image = $5 WHERE id = $6"
//			queryAddress = "UPDATE addresses SET text = $1, place_id = $2, ltd = $3, lng = $4 WHERE shop_id = $5"
//		)
//
//		// Init tx
//		tx, err := s.DB.Begin()
//		if err != nil {
//			//slog.Error("Error beginning tx", "File: ", "shop_repository.go", "Func: " ,"Update()","SubFunc: ", "DB.Begin()", "Err: %s", err.Error())
//			return err
//		}
//
//		// Update shop
//		_, err = tx.Exec(queryShop, shop.Name, shop.Email, shop.Phone, shop.Instagram, shop.Image.String, shopID)
//		if err != nil {
//			tx.Rollback()
//			//slog.Error("File: shop_repository.go | Func: Update() | SubFunc: tx.Exec() | Msg: Error updating shop | Err: %s", err.Error())
//			return err
//		}
//
//		// Update address
//		_, err = tx.Exec(queryAddress, shop.Address.Text, shop.Address.PlaceID, shop.Address.Ltd, shop.Address.Lng, shopID)
//		if err != nil {
//			tx.Rollback()
//			//slog.Error("File: shop_repository.go | Func: Update() | SubFunc: tx.Exec() | Msg: Error updating shop address | Err: %s", err.Error())
//			return err
//		}
//
//		// Confirm tx
//		err = tx.Commit()
//		if err != nil {
//			tx.Rollback()
//			//slog.Error("File: shop_repository.go | Func: Update() | SubFunc: tx.Exec() | Msg: Error committing tx | Err: %s", err.Error())
//			return err
//		}
//
//		return nil
//	}
//
//	func (s *ShopRepository) GetCategories(ctx context.Context, shopID int) ([]*entities.Category, error) {
//		query := `
//	           SELECT
//	               c.id, c.name, c.image
//	           FROM categories c
//	           WHERE c.shop_id = $1
//	           ORDER BY c.name
//	       `
//
//		rows, err := s.DB.Query(query, shopID)
//		if err != nil {
//			//slog.Error("File: shop_repository.go | Func: GetCategories() | SubFunc: DB.Query() | Msg: Error executing query | Err: %s", err.Error())
//			return nil, err
//		}
//		defer rows.Close()
//
//		// var categories []*entities.Category
//		categories := make([]*entities.Category, 0)
//		for rows.Next() {
//			category := &entities.Category{}
//			err := rows.Scan(&category.ID, &category.Name, &category.Image)
//			if err != nil {
//				//slog.Error("File: shop_repository.go | Func: GetCategories() | SubFunc: rows.Scan() | Msg: Error scanning row | Err: %s", err.Error())
//				return nil, err
//			}
//			categories = append(categories, category)
//		}
//
//		if err := rows.Err(); err != nil {
//			//slog.Error("File: shop_repository.go | Func: GetCategories() | SubFunc: rows.Err() | Msg: Error reading rows | Err: %s", err.Error())
//			return nil, err
//		}
//
//		return categories, nil
//	}
//
//	func (s *ShopRepository) GetProducts(ctx context.Context, shopID int) ([]*entities.Product, error) {
//		query := `
//			SELECT p.id, p.name, p.description, p.price, p.image, p.is_active,
//				   c.id AS c_id, c.name AS c_name, c.image AS c_image,
//				   COALESCE(o.id, 0) AS o_id, COALESCE(o.name, '') AS o_name, COALESCE(o.price, 0) AS o_price,
//				   COALESCE(v.id, 0) AS v_id, COALESCE(v.name, '') AS v_name
//			FROM public.products p
//			JOIN public.categories c ON p.category_id = c.id
//			LEFT JOIN public.options o ON p.id = o.product_id
//			LEFT JOIN public.variants v ON p.id = v.product_id
//			WHERE p.shop_id = $1
//		`
//
//		rows, err := s.DB.Query(query, shopID)
//		if err != nil {
//			return nil, err
//		}
//		defer rows.Close()
//
//		productsMap := make(map[int]*entities.Product)
//		for rows.Next() {
//			product := &entities.Product{}
//			category := &entities.Category{}
//			option := &entities.Option{}
//			variant := &entities.Variant{} // Nueva estructura para las variants
//
//			err := rows.Scan(
//				&product.ID, &product.Name, &product.Description, &product.Price, &product.Image, &product.IsActive,
//				&category.ID, &category.Name, &category.Image,
//				&option.ID, &option.Name, &option.Price,
//				&variant.ID, &variant.Name, // Nuevos campos para la variante
//			)
//			if err != nil {
//				return nil, err
//			}
//			product.Category = category
//
//			// Verificar si el production ya existe en el mapa
//			if existingProduct, exists := productsMap[product.ID]; exists {
//				// Si existe, solo agregar la opción y la variante si no son nulas
//				if option.ID != 0 {
//					existingProduct.Options = append(existingProduct.Options, option)
//				}
//				if variant.ID != 0 {
//					existingProduct.Variants = append(existingProduct.Variants, variant)
//				}
//			} else {
//				// Si no existe, agregar el production al mapa
//				if option.ID != 0 {
//					product.Options = append(product.Options, option)
//				}
//				if variant.ID != 0 {
//					product.Variants = append(product.Variants, variant)
//				}
//				productsMap[product.ID] = product
//			}
//		}
//
//		if err := rows.Err(); err != nil {
//			return nil, err
//		}
//
//		// Convertir el mapa en una lista
//		products := make([]*entities.Product, 0, len(productsMap))
//		for _, product := range productsMap {
//			products = append(products, product)
//		}
//
//		return products, nil
//	}
func NewShopRepository(dataBaseConnection DataBaseConnection) ports.ShopRepository {
	return &ShopSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}
