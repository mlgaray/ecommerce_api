package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Note: Uses 'json' variable from product_repository.go (same package)

// Shop repository log field constants
const (
	ShopRepositoryField        = "shop_repository"
	ShopGetByIDFunctionField   = "get_by_id"
	ShopGetBySlugFunctionField = "get_by_slug"
)

// shopQueryBase is the optimized query for loading a shop with all relations.
// Uses GROUP BY subqueries instead of LATERAL joins for 19x better performance.
// Planning: 4ms, Execution: 1ms (vs 78ms/17ms with LATERAL).
const shopQueryBase = `
	SELECT
		s.id, s.name, s.slug, s.email, s.phone, s.instagram, s.primary_color,
		COALESCE(img.images, '[]'::jsonb) AS images,
		CASE WHEN a.id IS NOT NULL THEN
			jsonb_build_object('id', a.id, 'name', a.name, 'place_id', a.place_id, 'lat', a.ltd, 'lng', a.lng)
		END AS address,
		COALESCE(pm.payment_methods, '[]'::jsonb) AS payment_methods,
		COALESCE(dm.delivery_methods, '[]'::jsonb) AS delivery_methods,
		COALESCE(os.operating_schedules, '[]'::jsonb) AS operating_schedules,
		CASE WHEN tz.id IS NOT NULL THEN
			jsonb_build_object('id', tz.id, 'name', tz.name, 'identifier', tz.identifier, 'utc_offset', tz.utc_offset)
		END AS timezone
	FROM shops s

	-- =========================
	-- Address (1-1) - Simple LEFT JOIN
	-- =========================
	LEFT JOIN addresses a ON a.shop_id = s.id

	-- =========================
	-- Timezone (1-1) - Simple LEFT JOIN
	-- =========================
	LEFT JOIN timezones tz ON tz.id = s.timezone_id

	-- =========================
	-- Images (aggregated by shop_id)
	-- =========================
	LEFT JOIN (
		SELECT
			i.shop_id,
			jsonb_agg(
				jsonb_build_object('id', i.id, 'url', i.url, 'type', i.type)
				ORDER BY i.type DESC
			) AS images
		FROM images i
		WHERE i.type IS NOT NULL
		GROUP BY i.shop_id
	) img ON img.shop_id = s.id

	-- =========================
	-- Payment Methods (aggregated with configs)
	-- =========================
	LEFT JOIN (
		SELECT
			spm.shop_id,
			jsonb_agg(
				jsonb_build_object(
					'id', pmt.id,
					'name', pmt.name,
					'code', pmt.code,
					'is_active', spm.is_active,
					'transfer_config', CASE WHEN tc.id IS NOT NULL THEN
						jsonb_build_object('id', tc.id, 'cbu', tc.cbu, 'cuil', tc.cuil, 'alias', tc.alias, 'owner_name', tc.owner_name)
					END,
					'mercadopago_config', CASE WHEN mc.id IS NOT NULL THEN
						jsonb_build_object('id', mc.id, 'access_token', mc.access_token, 'public_key', mc.public_key, 'user_id', mc.user_id)
					END
				) ORDER BY pmt.id
			) AS payment_methods
		FROM shop_payment_methods spm
		JOIN payment_methods pmt ON pmt.id = spm.payment_method_id
		LEFT JOIN transfer_configs tc ON tc.shop_payment_method_id = spm.id
		LEFT JOIN mercadopago_configs mc ON mc.shop_payment_method_id = spm.id
		GROUP BY spm.shop_id
	) pm ON pm.shop_id = s.id

	-- =========================
	-- Delivery Methods (aggregated with configs and zones)
	-- =========================
	LEFT JOIN (
		SELECT
			sdm.shop_id,
			jsonb_agg(
				jsonb_build_object(
					'id', dmt.id,
					'name', dmt.name,
					'code', dmt.code,
					'is_active', sdm.is_active,
					'delivery_config', CASE WHEN dc.id IS NOT NULL THEN
						jsonb_build_object('id', dc.id, 'fixed_price', dc.fixed_price)
					END,
					'pickup_config', CASE WHEN pc.id IS NOT NULL THEN
						jsonb_build_object('id', pc.id, 'address', pc.address, 'city', pc.city, 'province', pc.province, 'postal_code', pc.postal_code, 'instructions', pc.instructions)
					END,
					'delivery_zones', COALESCE(dz.zones, '[]'::jsonb)
				) ORDER BY dmt.id
			) AS delivery_methods
		FROM shop_delivery_methods sdm
		JOIN delivery_methods dmt ON dmt.id = sdm.delivery_method_id
		LEFT JOIN delivery_configs dc ON dc.shop_delivery_method_id = sdm.id
		LEFT JOIN pickup_configs pc ON pc.shop_delivery_method_id = sdm.id
		LEFT JOIN (
			SELECT
				dzn.shop_delivery_method_id,
				jsonb_agg(
					jsonb_build_object('id', dzn.id, 'name', dzn.name, 'price', dzn.price) ORDER BY dzn.id
				) AS zones
			FROM delivery_zones dzn
			GROUP BY dzn.shop_delivery_method_id
		) dz ON dz.shop_delivery_method_id = sdm.id
		GROUP BY sdm.shop_id
	) dm ON dm.shop_id = s.id

	-- =========================
	-- Operating Schedules (aggregated by shop_id)
	-- =========================
	LEFT JOIN (
		SELECT
			osch.shop_id,
			jsonb_agg(
				jsonb_build_object(
					'id', osch.id, 'day_of_week', osch.day_of_week,
					'open_time', osch.open_time::text, 'close_time', osch.close_time::text
				) ORDER BY osch.day_of_week, osch.open_time
			) AS operating_schedules
		FROM operating_schedules osch
		GROUP BY osch.shop_id
	) os ON os.shop_id = s.id
`

type ShopSQLRepository struct {
	db                 *sql.DB
	paymentMethodRepo  ports.PaymentMethodRepository
	deliveryMethodRepo ports.DeliveryMethodRepository
}

func (s *ShopSQLRepository) Create(ctx context.Context, userID int, shop *models.Shop) (*models.Shop, error) {
	// Extraer transacción del contexto si existe
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.createWithTx(ctx, tx, userID, shop)
	}

	// Si no hay transacción, usar conexión directa
	return s.createWithDB(ctx, userID, shop)
}

func (s *ShopSQLRepository) createWithTx(ctx context.Context, tx *sql.Tx, userID int, shop *models.Shop) (*models.Shop, error) {
	// Get default timezone ID (America/Buenos_Aires)
	var defaultTimezoneID *int
	err := tx.QueryRowContext(ctx, `SELECT id FROM timezones WHERE identifier = 'America/Buenos_Aires' LIMIT 1`).Scan(&defaultTimezoneID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	const query = `
		INSERT INTO shops (user_id, name, slug, email, phone, instagram, timezone_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var shopID int
	err = tx.QueryRowContext(ctx, query, userID, shop.Name, shop.Slug, shop.Email, shop.Phone, shop.Instagram, defaultTimezoneID).Scan(&shopID)
	if err != nil {
		return nil, s.handleInsertShopError(err, shop.Slug)
	}
	shop.ID = shopID

	// Crear payment methods para el shop (is_active = false)
	paymentMethods, err := s.paymentMethodRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	err = s.createPaymentMethodsWithTx(ctx, tx, shop.ID, paymentMethods)
	if err != nil {
		return nil, err
	}

	// Crear delivery methods para el shop (is_active = false)
	deliveryMethods, err := s.deliveryMethodRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	err = s.createDeliveryMethodsWithTx(ctx, tx, shop.ID, deliveryMethods)
	if err != nil {
		return nil, err
	}

	return shop, nil
}

func (s *ShopSQLRepository) createWithDB(ctx context.Context, userID int, shop *models.Shop) (*models.Shop, error) {
	// Get default timezone ID (America/Buenos_Aires)
	var defaultTimezoneID *int
	err := s.db.QueryRowContext(ctx, `SELECT id FROM timezones WHERE identifier = 'America/Buenos_Aires' LIMIT 1`).Scan(&defaultTimezoneID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	const query = `
		INSERT INTO shops (user_id, name, slug, email, phone, instagram, timezone_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var shopID int
	err = s.db.QueryRowContext(ctx, query, userID, shop.Name, shop.Slug, shop.Email, shop.Phone, shop.Instagram, defaultTimezoneID).Scan(&shopID)
	if err != nil {
		return nil, s.handleInsertShopError(err, shop.Slug)
	}
	shop.ID = shopID

	// Crear payment methods para el shop (is_active = false)
	paymentMethods, err := s.paymentMethodRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	err = s.createPaymentMethodsWithDB(ctx, shop.ID, paymentMethods)
	if err != nil {
		return nil, err
	}

	// Crear delivery methods para el shop (is_active = false)
	deliveryMethods, err := s.deliveryMethodRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	err = s.createDeliveryMethodsWithDB(ctx, shop.ID, deliveryMethods)
	if err != nil {
		return nil, err
	}

	return shop, nil
}

// handleInsertShopError translates PostgreSQL errors from shop INSERT to domain errors.
func (s *ShopSQLRepository) handleInsertShopError(err error, slug string) error {
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == PqErrCodeUniqueViolation && pqErr.Constraint == ConstraintShopSlugUnique {
			logs.WithFields(map[string]interface{}{
				"file":       ShopRepositoryField,
				"function":   "create",
				"constraint": pqErr.Constraint,
				"slug":       slug,
			}).Warn("Shop with slug already exists")
			return &errors.DuplicateRecordError{Message: errors.SlugAlreadyExists}
		}
	}
	logs.WithFields(map[string]interface{}{
		"file":     ShopRepositoryField,
		"function": "create",
		"slug":     slug,
		"error":    err.Error(),
	}).Error("Failed to create shop")
	return fmt.Errorf("database operation failed")
}

// GetByID returns a shop with all its related entities loaded using an optimized query.
// Uses GROUP BY subqueries instead of LATERAL joins for better performance.
func (s *ShopSQLRepository) GetByID(ctx context.Context, shopID int) (*models.Shop, error) {
	query := shopQueryBase + ` WHERE s.id = $1`

	shop := &models.Shop{}
	var primaryColor sql.NullString
	var imagesJSON, addressJSON, paymentMethodsJSON, deliveryMethodsJSON, schedulesJSON, timezoneJSON []byte

	err := s.db.QueryRowContext(ctx, query, shopID).Scan(
		&shop.ID,
		&shop.Name,
		&shop.Slug,
		&shop.Email,
		&shop.Phone,
		&shop.Instagram,
		&primaryColor,
		&imagesJSON,
		&addressJSON,
		&paymentMethodsJSON,
		&deliveryMethodsJSON,
		&schedulesJSON,
		&timezoneJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logs.WithFields(map[string]interface{}{
				"file":     ShopRepositoryField,
				"function": ShopGetByIDFunctionField,
				"shop_id":  shopID,
			}).Warn("Shop not found")
			return nil, &errors.RecordNotFoundError{Message: errors.ShopNotFound}
		}
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": ShopGetByIDFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to get shop by ID")
		return nil, fmt.Errorf("database operation failed")
	}

	// Convert nullable fields
	if primaryColor.Valid {
		shop.PrimaryColor = &primaryColor.String
	}

	// Unmarshal JSON fields
	if err := s.unmarshalShopRelations(shop, imagesJSON, addressJSON, paymentMethodsJSON, deliveryMethodsJSON, schedulesJSON, timezoneJSON); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": ShopGetByIDFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to unmarshal shop relations")
		return nil, fmt.Errorf("database operation failed")
	}

	return shop, nil
}

func (s *ShopSQLRepository) unmarshalShopRelations(
	shop *models.Shop,
	imagesJSON, addressJSON, paymentMethodsJSON, deliveryMethodsJSON, schedulesJSON, timezoneJSON []byte,
) error {
	if err := json.Unmarshal(imagesJSON, &shop.Images); err != nil {
		return fmt.Errorf("unmarshal images: %w", err)
	}

	if err := s.unmarshalNullableAddress(shop, addressJSON); err != nil {
		return err
	}

	if err := json.Unmarshal(paymentMethodsJSON, &shop.PaymentMethods); err != nil {
		return fmt.Errorf("unmarshal payment_methods: %w", err)
	}

	if err := json.Unmarshal(deliveryMethodsJSON, &shop.DeliveryMethods); err != nil {
		return fmt.Errorf("unmarshal delivery_methods: %w", err)
	}

	if err := json.Unmarshal(schedulesJSON, &shop.OperatingSchedules); err != nil {
		return fmt.Errorf("unmarshal operating_schedules: %w", err)
	}

	return s.unmarshalNullableTimezone(shop, timezoneJSON)
}

func (s *ShopSQLRepository) unmarshalNullableAddress(shop *models.Shop, addressJSON []byte) error {
	if len(addressJSON) > 0 && string(addressJSON) != JSONNull {
		shop.Address = &models.Address{}
		if err := json.Unmarshal(addressJSON, shop.Address); err != nil {
			return fmt.Errorf("unmarshal address: %w", err)
		}
	}
	return nil
}

func (s *ShopSQLRepository) unmarshalNullableTimezone(shop *models.Shop, timezoneJSON []byte) error {
	if len(timezoneJSON) > 0 && string(timezoneJSON) != JSONNull {
		shop.Timezone = &models.Timezone{}
		if err := json.Unmarshal(timezoneJSON, shop.Timezone); err != nil {
			return fmt.Errorf("unmarshal timezone: %w", err)
		}
	}
	return nil
}

// SlugExists checks if a shop with the given slug exists (case-insensitive).
// Uses lower() to match the unique index idx_shops_slug_unique on lower(slug).
// Lightweight query with no JOINs — used for slug availability checks.
func (s *ShopSQLRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM shops WHERE lower(slug) = lower($1))`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": "slug_exists",
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Failed to check slug existence")
		return false, fmt.Errorf("database operation failed")
	}

	return exists, nil
}

// GetBySlug returns a shop with all its related entities loaded using an optimized query.
// Uses GROUP BY subqueries instead of LATERAL joins for better performance.
// Used for public store view (no authentication required).
func (s *ShopSQLRepository) GetBySlug(ctx context.Context, slug string) (*models.Shop, error) {
	query := shopQueryBase + ` WHERE s.slug = $1`

	shop := &models.Shop{}
	var primaryColor sql.NullString
	var imagesJSON, addressJSON, paymentMethodsJSON, deliveryMethodsJSON, schedulesJSON, timezoneJSON []byte

	err := s.db.QueryRowContext(ctx, query, slug).Scan(
		&shop.ID,
		&shop.Name,
		&shop.Slug,
		&shop.Email,
		&shop.Phone,
		&shop.Instagram,
		&primaryColor,
		&imagesJSON,
		&addressJSON,
		&paymentMethodsJSON,
		&deliveryMethodsJSON,
		&schedulesJSON,
		&timezoneJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logs.WithFields(map[string]interface{}{
				"file":     ShopRepositoryField,
				"function": ShopGetBySlugFunctionField,
				"slug":     slug,
			}).Warn("Shop not found by slug")
			return nil, &errors.RecordNotFoundError{Message: errors.StoreNotFound}
		}
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": ShopGetBySlugFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Failed to get shop by slug")
		return nil, fmt.Errorf("database operation failed")
	}

	// Convert nullable fields
	if primaryColor.Valid {
		shop.PrimaryColor = &primaryColor.String
	}

	// Unmarshal JSON fields
	if err := s.unmarshalShopRelations(shop, imagesJSON, addressJSON, paymentMethodsJSON, deliveryMethodsJSON, schedulesJSON, timezoneJSON); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": ShopGetBySlugFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Failed to unmarshal shop relations")
		return nil, fmt.Errorf("database operation failed")
	}

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
//		err := row.ScanField(
//			&shop.ID, &shop.Name, &shop.Slug, &shop.Email, &shop.Phone, &shop.Instagram, &shop.Logo,
//			&shop.Address.ID, &shop.Address.Text, &shop.Address.PlaceID, &shop.Address.Ltd, &shop.Address.Lng,
//		)
//		if err != nil {
//			if err == sql.ErrNoRows {
//				//slog.Error("File: shop_repository.go | Func: GetByID() | SubFunc: row.ScanField() | Msg: Shop with id %d no found | Err: %s", shopID, err.Error())
//				return nil, errors.New("Shop not found")
//			}
//			//slog.Error("File: shop_repository.go | Func: GetByID() | SubFunc: row.ScanField() | Msg: Error scanning row | Err: %s", shopID, err.Error())
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
//			err := rows.ScanField(
//				&shop.ID, &shop.Name, &shop.Slug, &shop.Email, &shop.Phone, &shop.Instagram, &shop.Logo,
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
//			err := rows.ScanField(&category.ID, &category.Name, &category.Image)
//			if err != nil {
//				//slog.Error("File: shop_repository.go | Func: GetCategories() | SubFunc: rows.ScanField() | Msg: Error scanning row | Err: %s", err.Error())
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
//			err := rows.ScanField(
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
//
// GetShopsByUserID returns all shops owned by a user.
// Used during authentication to include shop IDs in JWT token.
func (s *ShopSQLRepository) GetShopsByUserID(ctx context.Context, userID int) ([]*models.Shop, error) {
	const query = `
		SELECT id, name, slug, email, phone, instagram
		FROM shops
		WHERE user_id = $1
		ORDER BY id
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	shops := make([]*models.Shop, 0)
	for rows.Next() {
		shop := &models.Shop{}
		err := rows.Scan(
			&shop.ID,
			&shop.Name,
			&shop.Slug,
			&shop.Email,
			&shop.Phone,
			&shop.Instagram,
		)
		if err != nil {
			return nil, fmt.Errorf("database operation failed")
		}
		shops = append(shops, shop)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database operation failed")
	}

	return shops, nil
}

// ==================== Payment Methods (internal) ====================

func (s *ShopSQLRepository) createPaymentMethodsWithTx(ctx context.Context, tx *sql.Tx, shopID int, methods []*models.PaymentMethod) error {
	const query = `INSERT INTO shop_payment_methods (shop_id, payment_method_id, is_active) VALUES ($1, $2, false)`

	for _, method := range methods {
		_, err := tx.ExecContext(ctx, query, shopID, method.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ShopSQLRepository) createPaymentMethodsWithDB(ctx context.Context, shopID int, methods []*models.PaymentMethod) error {
	const query = `INSERT INTO shop_payment_methods (shop_id, payment_method_id, is_active) VALUES ($1, $2, false)`

	for _, method := range methods {
		_, err := s.db.ExecContext(ctx, query, shopID, method.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// ==================== Delivery Methods (internal) ====================

func (s *ShopSQLRepository) createDeliveryMethodsWithTx(ctx context.Context, tx *sql.Tx, shopID int, methods []*models.DeliveryMethod) error {
	const query = `INSERT INTO shop_delivery_methods (shop_id, delivery_method_id, is_active) VALUES ($1, $2, false)`

	for _, method := range methods {
		_, err := tx.ExecContext(ctx, query, shopID, method.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ShopSQLRepository) createDeliveryMethodsWithDB(ctx context.Context, shopID int, methods []*models.DeliveryMethod) error {
	const query = `INSERT INTO shop_delivery_methods (shop_id, delivery_method_id, is_active) VALUES ($1, $2, false)`

	for _, method := range methods {
		_, err := s.db.ExecContext(ctx, query, shopID, method.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// ==================== Operating Schedules ====================

// GetOperatingSchedules returns all operating schedules for a shop, ordered by day_of_week and open_time.
func (s *ShopSQLRepository) GetOperatingSchedules(ctx context.Context, shopID int) ([]*models.OperatingSchedule, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.getOperatingSchedulesWithTx(ctx, tx, shopID)
	}
	return s.getOperatingSchedulesWithDB(ctx, shopID)
}

func (s *ShopSQLRepository) getOperatingSchedulesWithTx(ctx context.Context, tx *sql.Tx, shopID int) ([]*models.OperatingSchedule, error) {
	const query = `
		SELECT id, shop_id, day_of_week, open_time, close_time, created_at
		FROM operating_schedules
		WHERE shop_id = $1
		ORDER BY day_of_week, open_time
	`

	rows, err := tx.QueryContext(ctx, query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanOperatingSchedules(rows)
}

func (s *ShopSQLRepository) getOperatingSchedulesWithDB(ctx context.Context, shopID int) ([]*models.OperatingSchedule, error) {
	const query = `
		SELECT id, shop_id, day_of_week, open_time, close_time, created_at
		FROM operating_schedules
		WHERE shop_id = $1
		ORDER BY day_of_week, open_time
	`

	rows, err := s.db.QueryContext(ctx, query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanOperatingSchedules(rows)
}

func (s *ShopSQLRepository) scanOperatingSchedules(rows *sql.Rows) ([]*models.OperatingSchedule, error) {
	var schedules []*models.OperatingSchedule
	for rows.Next() {
		var os models.OperatingSchedule
		var openTime, closeTime time.Time
		var shopID int          // DB column, not needed in domain model
		var createdAt time.Time // DB column, not needed in domain model

		err := rows.Scan(
			&os.ID, &shopID, &os.DayOfWeek,
			&openTime, &closeTime, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert TIME fields to "HH:MM" string format
		os.OpenTime = openTime.Format("15:04")
		os.CloseTime = closeTime.Format("15:04")

		schedules = append(schedules, &os)
	}
	return schedules, rows.Err()
}

// SetOperatingSchedules replaces all operating schedules for a shop.
// Deletes existing schedules and inserts new ones in a transaction.
func (s *ShopSQLRepository) SetOperatingSchedules(ctx context.Context, shopID int, schedules []*models.OperatingSchedule) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Delete existing schedules
	const deleteQuery = `DELETE FROM operating_schedules WHERE shop_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, shopID)
	if err != nil {
		return fmt.Errorf("failed to delete existing schedules: %w", err)
	}

	// Insert new schedules
	const insertQuery = `
		INSERT INTO operating_schedules (shop_id, day_of_week, open_time, close_time)
		VALUES ($1, $2, $3, $4)
	`
	for _, schedule := range schedules {
		_, err = tx.ExecContext(ctx, insertQuery, shopID, schedule.DayOfWeek, schedule.OpenTime, schedule.CloseTime)
		if err != nil {
			return fmt.Errorf("failed to insert schedule: %w", err)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// IsShopOpen checks if the shop is open at the given time.
// Returns true if there is an operating schedule that covers the given time.
func (s *ShopSQLRepository) IsShopOpen(ctx context.Context, shopID int, checkTime time.Time) (bool, error) {
	// Get day of week (0=Sunday, 6=Saturday)
	dayOfWeek := int(checkTime.Weekday())
	timeStr := checkTime.Format("15:04")

	const query = `
		SELECT EXISTS (
			SELECT 1 FROM operating_schedules
			WHERE shop_id = $1
			  AND day_of_week = $2
			  AND open_time <= $3::TIME
			  AND close_time >= $3::TIME
		)
	`

	var isOpen bool
	err := s.db.QueryRowContext(ctx, query, shopID, dayOfWeek, timeStr).Scan(&isOpen)
	if err != nil {
		return false, fmt.Errorf("failed to check if shop is open: %w", err)
	}

	return isOpen, nil
}

// Update updates a shop and all its relations using a stored procedure.
// Returns array of storage_refs of deleted images for Cloudinary cleanup.
//
//nolint:gocyclo // Complexity is intentional for readability - JSON serialization and error handling
func (s *ShopSQLRepository) Update(ctx context.Context, shopID int, shop *models.Shop) ([]string, error) {
	startTime := time.Now()

	// Serialize images to JSONB
	imagesJSON, err := json.Marshal(shop.Images)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": "update",
			"sub_func": "marshal_images",
			"error":    err.Error(),
		}).Error("Failed to marshal images")
		return nil, fmt.Errorf("database operation failed")
	}

	// Serialize address to JSONB (can be nil)
	var addressJSON []byte
	if shop.Address != nil {
		addressJSON, err = json.Marshal(shop.Address)
		if err != nil {
			logs.WithFields(map[string]interface{}{
				"file":     ShopRepositoryField,
				"function": "update",
				"sub_func": "marshal_address",
				"error":    err.Error(),
			}).Error("Failed to marshal address")
			return nil, fmt.Errorf("database operation failed")
		}
	}

	// Serialize payment methods to JSONB
	paymentMethodsJSON, err := json.Marshal(shop.PaymentMethods)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": "update",
			"sub_func": "marshal_payment_methods",
			"error":    err.Error(),
		}).Error("Failed to marshal payment methods")
		return nil, fmt.Errorf("database operation failed")
	}

	// Serialize delivery methods to JSONB
	deliveryMethodsJSON, err := json.Marshal(shop.DeliveryMethods)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": "update",
			"sub_func": "marshal_delivery_methods",
			"error":    err.Error(),
		}).Error("Failed to marshal delivery methods")
		return nil, fmt.Errorf("database operation failed")
	}

	// Serialize operating schedules to JSONB
	schedulesJSON, err := json.Marshal(shop.OperatingSchedules)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": "update",
			"sub_func": "marshal_schedules",
			"error":    err.Error(),
		}).Error("Failed to marshal operating schedules")
		return nil, fmt.Errorf("database operation failed")
	}

	// Call stored procedure
	var deletedRefs []string
	err = s.db.QueryRowContext(ctx, `
		SELECT update_shop($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		shopID,
		shop.Name,
		shop.Slug,
		shop.Email,
		shop.Phone,
		shop.Instagram,
		shop.PrimaryColor,
		imagesJSON,
		addressJSON,
		paymentMethodsJSON,
		deliveryMethodsJSON,
		schedulesJSON,
	).Scan(pq.Array(&deletedRefs))

	if err != nil {
		// Check if it's a PostgreSQL error from the stored procedure
		if pqErr, ok := err.(*pq.Error); ok {
			// Check for "Shop not found" (P0002)
			if pqErr.Code == PqErrCodeNoDataFound {
				logs.WithFields(map[string]interface{}{
					"file":     ShopRepositoryField,
					"function": "update",
					"shop_id":  shopID,
				}).Warn("Shop not found")
				return nil, &errors.RecordNotFoundError{Message: errors.ShopNotFound}
			}

			// Unique violation (e.g., duplicate slug/email)
			if pqErr.Code == PqErrCodeUniqueViolation {
				logs.WithFields(map[string]interface{}{
					"file":       ShopRepositoryField,
					"function":   "update",
					"shop_id":    shopID,
					"pg_code":    pqErr.Code,
					"pg_message": pqErr.Message,
				}).Warn("Unique constraint violation")
				return nil, &errors.ValidationError{Message: "unique constraint violation: " + pqErr.Message}
			}

			logs.WithFields(map[string]interface{}{
				"file":       ShopRepositoryField,
				"function":   "update",
				"shop_id":    shopID,
				"pg_code":    pqErr.Code,
				"pg_message": pqErr.Message,
			}).Error("PostgreSQL error")

			return nil, fmt.Errorf("database operation failed")
		}

		logs.WithFields(map[string]interface{}{
			"file":     ShopRepositoryField,
			"function": "update",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to update shop")
		return nil, fmt.Errorf("database operation failed")
	}

	logs.WithFields(map[string]interface{}{
		"file":              ShopRepositoryField,
		"function":          "update",
		"shop_id":           shopID,
		"deleted_refs":      len(deletedRefs),
		"total_duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Shop update completed (stored procedure)")

	return deletedRefs, nil
}

func NewShopRepository(
	dataBaseConnection DataBaseConnection,
	paymentMethodRepo ports.PaymentMethodRepository,
	deliveryMethodRepo ports.DeliveryMethodRepository,
) ports.ShopRepository {
	return &ShopSQLRepository{
		db:                 dataBaseConnection.Connect(),
		paymentMethodRepo:  paymentMethodRepo,
		deliveryMethodRepo: deliveryMethodRepo,
	}
}
