package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type ShopSQLRepository struct {
	db                 *sql.DB
	paymentMethodRepo  ports.PaymentMethodRepository
	deliveryMethodRepo ports.DeliveryMethodRepository
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
//			&shop.ID, &shop.Name, &shop.Slug, &shop.Email, &shop.Phone, &shop.Instagram, &shop.Image,
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
func (s *ShopSQLRepository) GetShopsByUserID(ctx context.Context, userID int) ([]*models.Shop, error) {
	const query = `
		SELECT id, name, slug, email, phone, instagram, image
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
		shop := &models.Shop{UserID: userID}
		err := rows.Scan(
			&shop.ID,
			&shop.Name,
			&shop.Slug,
			&shop.Email,
			&shop.Phone,
			&shop.Instagram,
			&shop.Image,
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

func (s *ShopSQLRepository) GetPaymentMethods(ctx context.Context, shopID int) ([]*models.ShopPaymentMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.getPaymentMethodsWithTx(ctx, tx, shopID)
	}
	return s.getPaymentMethodsWithDB(ctx, shopID)
}

func (s *ShopSQLRepository) getPaymentMethodsWithTx(ctx context.Context, tx *sql.Tx, shopID int) ([]*models.ShopPaymentMethod, error) {
	const query = `
		SELECT
			spm.id, spm.shop_id, spm.payment_method_id, spm.is_active,
			pm.id, pm.name, pm.code, pm.description, pm.is_active
		FROM shop_payment_methods spm
		JOIN payment_methods pm ON spm.payment_method_id = pm.id
		WHERE spm.shop_id = $1
		ORDER BY pm.id
	`

	rows, err := tx.QueryContext(ctx, query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanShopPaymentMethods(rows)
}

func (s *ShopSQLRepository) getPaymentMethodsWithDB(ctx context.Context, shopID int) ([]*models.ShopPaymentMethod, error) {
	const query = `
		SELECT
			spm.id, spm.shop_id, spm.payment_method_id, spm.is_active,
			pm.id, pm.name, pm.code, pm.description, pm.is_active
		FROM shop_payment_methods spm
		JOIN payment_methods pm ON spm.payment_method_id = pm.id
		WHERE spm.shop_id = $1
		ORDER BY pm.id
	`

	rows, err := s.db.QueryContext(ctx, query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanShopPaymentMethods(rows)
}

func (s *ShopSQLRepository) scanShopPaymentMethods(rows *sql.Rows) ([]*models.ShopPaymentMethod, error) {
	var methods []*models.ShopPaymentMethod
	for rows.Next() {
		var spm models.ShopPaymentMethod
		var pm models.PaymentMethod
		var pmDescription sql.NullString

		err := rows.Scan(
			&spm.ID, &spm.ShopID, &spm.PaymentMethodID, &spm.IsActive,
			&pm.ID, &pm.Name, &pm.Code, &pmDescription, &pm.IsActive,
		)
		if err != nil {
			return nil, err
		}

		if pmDescription.Valid {
			pm.Description = pmDescription.String
		}
		spm.PaymentMethod = &pm
		methods = append(methods, &spm)
	}
	return methods, rows.Err()
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

func (s *ShopSQLRepository) GetDeliveryMethods(ctx context.Context, shopID int) ([]*models.ShopDeliveryMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return s.getDeliveryMethodsWithTx(ctx, tx, shopID)
	}
	return s.getDeliveryMethodsWithDB(ctx, shopID)
}

func (s *ShopSQLRepository) getDeliveryMethodsWithTx(ctx context.Context, tx *sql.Tx, shopID int) ([]*models.ShopDeliveryMethod, error) {
	const query = `
		SELECT
			sdm.id, sdm.shop_id, sdm.delivery_method_id, sdm.is_active,
			dm.id, dm.name, dm.code, dm.description, dm.is_active
		FROM shop_delivery_methods sdm
		JOIN delivery_methods dm ON sdm.delivery_method_id = dm.id
		WHERE sdm.shop_id = $1
		ORDER BY dm.id
	`

	rows, err := tx.QueryContext(ctx, query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanShopDeliveryMethods(rows)
}

func (s *ShopSQLRepository) getDeliveryMethodsWithDB(ctx context.Context, shopID int) ([]*models.ShopDeliveryMethod, error) {
	const query = `
		SELECT
			sdm.id, sdm.shop_id, sdm.delivery_method_id, sdm.is_active,
			dm.id, dm.name, dm.code, dm.description, dm.is_active
		FROM shop_delivery_methods sdm
		JOIN delivery_methods dm ON sdm.delivery_method_id = dm.id
		WHERE sdm.shop_id = $1
		ORDER BY dm.id
	`

	rows, err := s.db.QueryContext(ctx, query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanShopDeliveryMethods(rows)
}

func (s *ShopSQLRepository) scanShopDeliveryMethods(rows *sql.Rows) ([]*models.ShopDeliveryMethod, error) {
	var methods []*models.ShopDeliveryMethod
	for rows.Next() {
		var sdm models.ShopDeliveryMethod
		var dm models.DeliveryMethod
		var dmDescription sql.NullString

		err := rows.Scan(
			&sdm.ID, &sdm.ShopID, &sdm.DeliveryMethodID, &sdm.IsActive,
			&dm.ID, &dm.Name, &dm.Code, &dmDescription, &dm.IsActive,
		)
		if err != nil {
			return nil, err
		}

		if dmDescription.Valid {
			dm.Description = dmDescription.String
		}
		sdm.DeliveryMethod = &dm
		methods = append(methods, &sdm)
	}
	return methods, rows.Err()
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

		err := rows.Scan(
			&os.ID, &os.ShopID, &os.DayOfWeek,
			&openTime, &closeTime, &os.CreatedAt,
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
