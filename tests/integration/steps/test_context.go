package steps

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/auth"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/category"
	couponuc "github.com/mlgaray/ecommerce_api/internal/application/usecases/coupon"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/metrics"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/order"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/product"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/shop"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/store"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	authBcrypt "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/bcrypt"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/jwt"
	authhttp "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/middleware"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/repositories/postgresql"
	"github.com/mlgaray/ecommerce_api/tests/integration/stubs"
)

// mockDataBaseConnection implements postgresql.DataBaseConnection for testing
type mockDataBaseConnection struct {
	db *sql.DB
}

func (m *mockDataBaseConnection) Connect() *sql.DB {
	return m.db
}

// TestContext contiene todo el estado compartido entre tests
type TestContext struct {
	mu sync.Mutex // Protects all fields below

	// HTTP
	app      *fx.App
	server   *httptest.Server
	response *http.Response

	// Generic request/response data (reutilizable para cualquier endpoint)
	requestBody  interface{} // Puede ser cualquier tipo de request
	responseBody interface{} // Puede ser cualquier tipo de response
	queryParams  map[string]string
	pathParams   map[string]string

	// Messages
	successMessage string
	errorMessage   string

	// Product-specific fields (para multipart/form-data)
	productImages    [][]byte
	invalidImageType bool

	// Authentication
	authToken string // JWT token for authenticated requests

	// Test control
	scenario string

	// SQL Mock
	mockDB      *sql.DB
	mockSQLMock sqlmock.Sqlmock
}

// Global test context instance with mutex for thread safety
var (
	testCtx   *TestContext
	testCtxMu sync.Mutex
)

// GetTestContext returns the current test context (thread-safe)
func GetTestContext() *TestContext {
	testCtxMu.Lock()
	defer testCtxMu.Unlock()
	if testCtx == nil {
		testCtx = &TestContext{}
	}
	return testCtx
}

// Reset clears all test context data
func (ctx *TestContext) Reset() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.app = nil
	ctx.server = nil
	ctx.response = nil
	ctx.requestBody = nil
	ctx.responseBody = nil
	ctx.queryParams = nil
	ctx.pathParams = nil
	ctx.successMessage = ""
	ctx.errorMessage = ""
	ctx.productImages = nil
	ctx.invalidImageType = false
	ctx.authToken = ""
	ctx.scenario = ""

	// Close existing resources
	if ctx.mockDB != nil {
		ctx.mockDB.Close()
		ctx.mockDB = nil
	}
	ctx.mockSQLMock = nil

	if ctx.server != nil {
		ctx.server.Close()
		ctx.server = nil
	}

	if ctx.app != nil {
		if err := ctx.app.Stop(context.Background()); err != nil {
			// TODO: Log error but continue cleanup
			_ = err
		}
		ctx.app = nil
	}
}

// SetupTestApp initializes the test application with mocked dependencies
func (ctx *TestContext) SetupTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order
	sqlMock.MatchExpectationsInOrder(false)

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide real services with interface annotations
			fx.Annotate(services.NewUserService, fx.As(new(ports.UserService))),
			fx.Annotate(authBcrypt.NewAuthService, fx.As(new(ports.AuthService))),
			fx.Annotate(services.NewSignupService, fx.As(new(ports.SignUpService))),
			fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),
			fx.Annotate(postgresql.NewUserRepository, fx.As(new(ports.UserRepository))),
			fx.Annotate(postgresql.NewRoleRepository, fx.As(new(ports.RoleRepository))),
			fx.Annotate(postgresql.NewPaymentMethodRepository, fx.As(new(ports.PaymentMethodRepository))),
			fx.Annotate(postgresql.NewDeliveryMethodRepository, fx.As(new(ports.DeliveryMethodRepository))),
			fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),
			fx.Annotate(postgresql.NewStaffRepository, fx.As(new(ports.StaffRepository))),
			fx.Annotate(stubs.NewAssetService, fx.As(new(ports.AssetService))),
			fx.Annotate(services.NewStaffService, fx.As(new(ports.StaffService))),
			fx.Annotate(postgresql.NewPermissionRepository, fx.As(new(ports.PermissionRepository))),
			services.NewPermissionService,
			func(svc *services.PermissionServiceImpl) ports.PermissionService { return svc },
			fx.Annotate(postgresql.NewSignupRepository, fx.As(new(ports.SignupRepository))),

			// Provide use cases
			fx.Annotate(auth.NewSignInUseCase, fx.As(new(ports.SignInUseCase))),
			fx.Annotate(auth.NewSignUpUseCase, fx.As(new(ports.SignUpUseCase))),

			// Provide handlers
			authhttp.NewAuthHandler,
		),
		fx.Invoke(func(handler *authhttp.AuthHandler) {
			// Create HTTP router and server
			router := mux.NewRouter()
			router.HandleFunc("/auth/signin", handler.SignIn).Methods("POST")
			router.HandleFunc("/auth/signup", handler.SignUp).Methods("POST")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// GenerateTestToken creates a JWT token for testing with default test user
func (ctx *TestContext) GenerateTestToken() error {
	tokenService := jwt.NewTokenService()
	testUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	testShopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}}

	token, err := tokenService.Generate(context.Background(), testUser, testShopRoles)
	if err != nil {
		return err
	}
	ctx.authToken = token
	return nil
}

// SetupProductTestApp initializes the test application for product tests
func (ctx *TestContext) SetupProductTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order (use case runs COUNT and SELECT in parallel)
	sqlMock.MatchExpectationsInOrder(false)

	// Create TokenService for auth middleware
	tokenService := jwt.NewTokenService()

	// Generate test token for authenticated requests
	testUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	testShopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}}
	ctx.authToken, err = tokenService.Generate(context.Background(), testUser, testShopRoles)
	if err != nil {
		return err
	}

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide TokenService for auth middleware
			fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),

			// Provide auth middleware
			middleware.NewAuthMiddleware,

			// Provide product dependencies (following dependency order)
			// 1. Repository first (no dependencies)
			fx.Annotate(postgresql.NewProductRepository, fx.As(new(ports.ProductRepository))),
			// 2. AssetService stub for testing (no external dependencies)
			fx.Annotate(stubs.NewAssetService, fx.As(new(ports.AssetService))),
			// 3. Service depends on Repository and AssetService
			fx.Annotate(services.NewProductService, fx.As(new(ports.ProductService))),

			// Provide pagination service
			fx.Annotate(
				services.NewPaginationService[*models.Product],
				fx.As(new(ports.PaginationService[*models.Product])),
			),

			// Provide use cases (depend on Services, NOT repositories)
			fx.Annotate(product.NewCreateProductUseCase, fx.As(new(ports.CreateProductUseCase))),
			fx.Annotate(product.NewGetAllByShopIDWithFiltersUseCase, fx.As(new(ports.GetAllByShopIDWithFiltersUseCase))),
			fx.Annotate(product.NewGetByIDUseCase, fx.As(new(ports.GetByIDUseCase))),
			fx.Annotate(product.NewUpdateProductUseCase, fx.As(new(ports.UpdateProductUseCase))),
			fx.Annotate(product.NewDeleteProductUseCase, fx.As(new(ports.DeleteProductUseCase))),

			// Provide handler
			authhttp.NewProductHandler,
		),
		fx.Invoke(func(handler *authhttp.ProductHandler, authMiddleware *middleware.AuthMiddleware) {
			// Create HTTP router and server
			router := mux.NewRouter()

			// Protected routes (require auth) - using REAL auth middleware
			protectedProducts := router.PathPrefix("/products").Subrouter()
			protectedProducts.Use(authMiddleware.Authenticate)
			protectedProducts.HandleFunc("", handler.Create).Methods("POST")
			protectedProducts.HandleFunc("/{product_id}", handler.GetByID).Methods("GET")
			protectedProducts.HandleFunc("/{product_id}", handler.Update).Methods("PUT")
			protectedProducts.HandleFunc("/{product_id}", handler.Delete).Methods("DELETE")

			// Public routes (no auth required)
			router.HandleFunc("/shops/{shop_id}/products", handler.GetAllByShopIDWithFilters).Methods("GET")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// SetupCategoryTestApp initializes the test application for category tests
func (ctx *TestContext) SetupCategoryTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order (use case runs COUNT and SELECT in parallel)
	sqlMock.MatchExpectationsInOrder(false)

	// Create TokenService for auth middleware
	tokenService := jwt.NewTokenService()

	// Generate test token for authenticated requests
	testUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	testShopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}}
	ctx.authToken, err = tokenService.Generate(context.Background(), testUser, testShopRoles)
	if err != nil {
		return err
	}

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide TokenService for auth middleware
			fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),

			// Provide auth middleware
			middleware.NewAuthMiddleware,

			// Provide category dependencies
			fx.Annotate(postgresql.NewCategoryRepository, fx.As(new(ports.CategoryRepository))),
			fx.Annotate(stubs.NewAssetService, fx.As(new(ports.AssetService))),
			fx.Annotate(services.NewCategoryService, fx.As(new(ports.CategoryService))),
			fx.Annotate(
				services.NewPaginationService[*models.Category],
				fx.As(new(ports.PaginationService[*models.Category])),
			),

			// Provide use cases
			fx.Annotate(category.NewCreateCategoryUseCase, fx.As(new(ports.CreateCategoryUseCase))),
			fx.Annotate(category.NewGetByIDUseCase, fx.As(new(ports.GetCategoryByIDUseCase))),
			fx.Annotate(category.NewGetAllByShopIDWithFiltersUseCase, fx.As(new(ports.GetAllCategoriesByShopIDWithFiltersUseCase))),
			fx.Annotate(category.NewUpdateCategoryUseCase, fx.As(new(ports.UpdateCategoryUseCase))),
			fx.Annotate(category.NewDeleteCategoryUseCase, fx.As(new(ports.DeleteCategoryUseCase))),

			// Provide handler
			authhttp.NewCategoryHandler,
		),
		fx.Invoke(func(handler *authhttp.CategoryHandler, authMiddleware *middleware.AuthMiddleware) {
			// Create HTTP router and server
			router := mux.NewRouter()

			// Protected routes (require auth) - using REAL auth middleware
			protectedCategories := router.PathPrefix("/categories").Subrouter()
			protectedCategories.Use(authMiddleware.Authenticate)
			protectedCategories.HandleFunc("", handler.Create).Methods("POST")
			protectedCategories.HandleFunc("/{category_id}", handler.GetByID).Methods("GET")
			protectedCategories.HandleFunc("/{category_id}", handler.Update).Methods("PUT")
			protectedCategories.HandleFunc("/{category_id}", handler.Delete).Methods("DELETE")

			// Public routes (no auth required)
			router.HandleFunc("/shops/{shop_id}/categories", handler.GetAllByShopIDWithFilters).Methods("GET")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// SetupShopTestApp initializes the test application for shop tests
func (ctx *TestContext) SetupShopTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order
	sqlMock.MatchExpectationsInOrder(false)

	// Create TokenService for auth middleware
	tokenService := jwt.NewTokenService()

	// Generate test token for authenticated requests
	// Include multiple shop IDs to support different test scenarios
	// - Shop 1, 2: standard test shops
	// - Shop 888: for "not found" tests (passes auth, but shop doesn't exist in DB)
	// - Shop 999: intentionally excluded for "not owned" tests
	testUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	testShopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}, {ShopID: 2, Role: "admin"}, {ShopID: 888, Role: "owner"}}
	ctx.authToken, err = tokenService.Generate(context.Background(), testUser, testShopRoles)
	if err != nil {
		return err
	}

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide TokenService for auth middleware
			fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),

			// Provide auth middleware
			middleware.NewAuthMiddleware,

			// Provide repository dependencies required by ShopRepository
			fx.Annotate(postgresql.NewPaymentMethodRepository, fx.As(new(ports.PaymentMethodRepository))),
			fx.Annotate(postgresql.NewDeliveryMethodRepository, fx.As(new(ports.DeliveryMethodRepository))),

			// Provide shop dependencies
			fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),
			fx.Annotate(stubs.NewAssetService, fx.As(new(ports.AssetService))),
			fx.Annotate(services.NewShopService, fx.As(new(ports.ShopService))),

			// Provide use cases
			fx.Annotate(shop.NewGetShopByIDUseCase, fx.As(new(ports.GetShopByIDUseCase))),
			fx.Annotate(shop.NewUpdateShopUseCase, fx.As(new(ports.UpdateShopUseCase))),

			// Provide handler
			fx.Annotate(authhttp.NewShopHandler, fx.As(new(ports.ShopHandler))),

			// Provide shop ownership middleware
			middleware.NewShopOwnershipMiddleware,
		),
		fx.Invoke(func(handler ports.ShopHandler, authMiddleware *middleware.AuthMiddleware, shopOwnershipMiddleware *middleware.ShopOwnershipMiddleware) {
			// Create HTTP router and server
			router := mux.NewRouter()

			// Protected routes (require auth + shop ownership) - using REAL middlewares
			protectedShops := router.PathPrefix("/shops").Subrouter()
			protectedShops.Use(authMiddleware.Authenticate)
			protectedShops.Use(shopOwnershipMiddleware.Authorize)
			protectedShops.HandleFunc("/{shop_id}", handler.GetByID).Methods("GET")
			protectedShops.HandleFunc("/{shop_id}", handler.Update).Methods("PUT")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// SetupStoreTestApp initializes the test application for store tests (public endpoints)
func (ctx *TestContext) SetupStoreTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order
	sqlMock.MatchExpectationsInOrder(false)

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide store dependencies (no auth required - public endpoints)
			fx.Annotate(postgresql.NewPaymentMethodRepository, fx.As(new(ports.PaymentMethodRepository))),
			fx.Annotate(postgresql.NewDeliveryMethodRepository, fx.As(new(ports.DeliveryMethodRepository))),
			fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),
			fx.Annotate(postgresql.NewStoreVisitRepository, fx.As(new(ports.StoreVisitRepository))),
			fx.Annotate(services.NewStoreService, fx.As(new(ports.StoreService))),

			// Provide category dependencies (for store categories endpoint)
			fx.Annotate(postgresql.NewCategoryRepository, fx.As(new(ports.CategoryRepository))),
			fx.Annotate(stubs.NewAssetService, fx.As(new(ports.AssetService))),
			fx.Annotate(services.NewCategoryService, fx.As(new(ports.CategoryService))),
			fx.Annotate(
				services.NewPaginationService[*models.Category],
				fx.As(new(ports.PaginationService[*models.Category])),
			),

			// Provide product dependencies (for store featured products endpoint)
			fx.Annotate(postgresql.NewProductRepository, fx.As(new(ports.ProductRepository))),
			fx.Annotate(services.NewProductService, fx.As(new(ports.ProductService))),
			fx.Annotate(
				services.NewPaginationService[*models.Product],
				fx.As(new(ports.PaginationService[*models.Product])),
			),

			// Provide coupon dependencies (for store coupon validation endpoint)
			fx.Annotate(postgresql.NewCouponRepository, fx.As(new(ports.CouponRepository))),
			fx.Annotate(services.NewCouponService, fx.As(new(ports.CouponService))),

			// Provide use cases
			fx.Annotate(store.NewCheckSlugAvailabilityUseCase, fx.As(new(ports.CheckSlugAvailabilityUseCase))),
			fx.Annotate(store.NewGetStoreBySlugUseCase, fx.As(new(ports.GetStoreBySlugUseCase))),
			fx.Annotate(store.NewGetStoreCategoriesUseCase, fx.As(new(ports.GetStoreCategoriesUseCase))),
			fx.Annotate(store.NewGetStoreProductsUseCase, fx.As(new(ports.GetStoreProductsUseCase))),
			fx.Annotate(store.NewGetStoreFeaturedProductsUseCase, fx.As(new(ports.GetStoreFeaturedProductsUseCase))),
			fx.Annotate(store.NewGetStoreProductByIDUseCase, fx.As(new(ports.GetStoreProductByIDUseCase))),
			fx.Annotate(couponuc.NewValidateStoreCouponUseCase, fx.As(new(ports.ValidateStoreCouponUseCase))),

			// Provide handler
			fx.Annotate(authhttp.NewStoreHandler, fx.As(new(ports.StoreHandler))),
		),
		fx.Invoke(func(handler ports.StoreHandler) {
			// Create HTTP router and server
			router := mux.NewRouter()

			// Public routes (no auth required) - Customer-facing store endpoints
			// Note: More specific routes must be registered before less specific ones
			router.HandleFunc("/stores/slugs/{slug}/check-availability", handler.CheckSlugAvailability).Methods("GET")
			router.HandleFunc("/stores/{slug}/products/featured", handler.GetFeaturedProducts).Methods("GET")
			router.HandleFunc("/stores/{slug}/products/{productId}", handler.GetProductByID).Methods("GET")
			router.HandleFunc("/stores/{slug}/products", handler.GetProducts).Methods("GET")
			router.HandleFunc("/stores/{slug}/categories", handler.GetCategories).Methods("GET")
			router.HandleFunc("/stores/{slug}", handler.GetBySlug).Methods("GET")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// SetupCreateOrderTestApp initializes the test application for create order tests (public endpoint)
func (ctx *TestContext) SetupCreateOrderTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order
	sqlMock.MatchExpectationsInOrder(false)

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide repository dependencies
			fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),
			fx.Annotate(postgresql.NewProductRepository, fx.As(new(ports.ProductRepository))),
			fx.Annotate(postgresql.NewPaymentMethodRepository, fx.As(new(ports.PaymentMethodRepository))),
			fx.Annotate(postgresql.NewDeliveryMethodRepository, fx.As(new(ports.DeliveryMethodRepository))),
			fx.Annotate(postgresql.NewOrderRepository, fx.As(new(ports.OrderRepository))),
			fx.Annotate(postgresql.NewCouponRepository, fx.As(new(ports.CouponRepository))),

			// Provide services
			fx.Annotate(postgresql.NewStoreVisitRepository, fx.As(new(ports.StoreVisitRepository))),
			fx.Annotate(services.NewStoreService, fx.As(new(ports.StoreService))),
			fx.Annotate(services.NewOrderService, fx.As(new(ports.OrderService))),
			fx.Annotate(services.NewCouponService, fx.As(new(ports.CouponService))),
			fx.Annotate(stubs.NewOrderEventNotifier, fx.As(new(ports.OrderEventNotifier))),
			fx.Annotate(stubs.NewAssetService, fx.As(new(ports.AssetService))),
			fx.Annotate(services.NewShopService, fx.As(new(ports.ShopService))),
			fx.Annotate(
				services.NewPaginationService[*models.Order],
				fx.As(new(ports.PaginationService[*models.Order])),
			),

			// Provide use cases (all 6 required by NewOrderHandler)
			fx.Annotate(order.NewCreateOrderUseCase, fx.As(new(ports.CreateOrderUseCase))),
			fx.Annotate(order.NewGetAllOrdersByShopIDUseCase, fx.As(new(ports.GetAllOrdersByShopIDUseCase))),
			fx.Annotate(order.NewGetOrderByIDUseCase, fx.As(new(ports.GetOrderByIDUseCase))),
			fx.Annotate(order.NewUpdateOrderStatusUseCase, fx.As(new(ports.UpdateOrderStatusUseCase))),
			fx.Annotate(order.NewUpdateOrderUseCase, fx.As(new(ports.UpdateOrderUseCase))),
			fx.Annotate(order.NewRemoveOrderCouponUseCase, fx.As(new(ports.RemoveOrderCouponUseCase))),

			// Provide handler
			fx.Annotate(authhttp.NewOrderHandler, fx.As(new(ports.OrderHandler))),
		),
		fx.Invoke(func(handler ports.OrderHandler) {
			// Create HTTP router and server
			router := mux.NewRouter()

			// Public route: POST /stores/{slug}/orders
			storeRouter := router.PathPrefix("/stores").Subrouter()
			storeRouter.HandleFunc("/{slug}/orders", handler.Create).Methods("POST")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// SetupOrderTestApp initializes the test application for order tests (protected endpoints)
func (ctx *TestContext) SetupOrderTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order
	sqlMock.MatchExpectationsInOrder(false)

	// Create TokenService for auth middleware
	tokenService := jwt.NewTokenService()

	// Generate test token for authenticated requests
	// - Shop 1, 2: standard test shops
	// - Shop 888: for "not found" tests (passes auth, but order doesn't exist in DB)
	// - Shop 999: intentionally excluded for "not owned" tests
	testUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	testShopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}, {ShopID: 2, Role: "admin"}, {ShopID: 888, Role: "owner"}}
	ctx.authToken, err = tokenService.Generate(context.Background(), testUser, testShopRoles)
	if err != nil {
		return err
	}

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide TokenService for auth middleware
			fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),

			// Provide auth middleware
			middleware.NewAuthMiddleware,

			// Provide repository dependencies
			fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),
			fx.Annotate(postgresql.NewProductRepository, fx.As(new(ports.ProductRepository))),
			fx.Annotate(postgresql.NewPaymentMethodRepository, fx.As(new(ports.PaymentMethodRepository))),
			fx.Annotate(postgresql.NewDeliveryMethodRepository, fx.As(new(ports.DeliveryMethodRepository))),
			fx.Annotate(postgresql.NewOrderRepository, fx.As(new(ports.OrderRepository))),
			fx.Annotate(postgresql.NewCouponRepository, fx.As(new(ports.CouponRepository))),

			// Provide services
			fx.Annotate(postgresql.NewStoreVisitRepository, fx.As(new(ports.StoreVisitRepository))),
			fx.Annotate(services.NewStoreService, fx.As(new(ports.StoreService))),
			fx.Annotate(services.NewOrderService, fx.As(new(ports.OrderService))),
			fx.Annotate(services.NewCouponService, fx.As(new(ports.CouponService))),
			fx.Annotate(stubs.NewOrderEventNotifier, fx.As(new(ports.OrderEventNotifier))),
			fx.Annotate(stubs.NewAssetService, fx.As(new(ports.AssetService))),
			fx.Annotate(services.NewShopService, fx.As(new(ports.ShopService))),
			fx.Annotate(
				services.NewPaginationService[*models.Order],
				fx.As(new(ports.PaginationService[*models.Order])),
			),

			// Provide use cases (all 6 required by NewOrderHandler)
			fx.Annotate(order.NewCreateOrderUseCase, fx.As(new(ports.CreateOrderUseCase))),
			fx.Annotate(order.NewGetAllOrdersByShopIDUseCase, fx.As(new(ports.GetAllOrdersByShopIDUseCase))),
			fx.Annotate(order.NewGetOrderByIDUseCase, fx.As(new(ports.GetOrderByIDUseCase))),
			fx.Annotate(order.NewUpdateOrderStatusUseCase, fx.As(new(ports.UpdateOrderStatusUseCase))),
			fx.Annotate(order.NewUpdateOrderUseCase, fx.As(new(ports.UpdateOrderUseCase))),
			fx.Annotate(order.NewRemoveOrderCouponUseCase, fx.As(new(ports.RemoveOrderCouponUseCase))),

			// Provide handler
			fx.Annotate(authhttp.NewOrderHandler, fx.As(new(ports.OrderHandler))),

			// Provide shop ownership middleware
			middleware.NewShopOwnershipMiddleware,
		),
		fx.Invoke(func(handler ports.OrderHandler, authMiddleware *middleware.AuthMiddleware, shopOwnershipMiddleware *middleware.ShopOwnershipMiddleware) {
			// Create HTTP router and server
			router := mux.NewRouter()

			// Protected routes (require auth + shop ownership)
			protected := router.PathPrefix("/shops").Subrouter()
			protected.Use(authMiddleware.Authenticate)
			protected.Use(shopOwnershipMiddleware.Authorize)
			protected.HandleFunc("/{shop_id}/orders/{order_id:[0-9]+}/status", handler.UpdateStatus).Methods("PATCH")
			protected.HandleFunc("/{shop_id}/orders/{order_id:[0-9]+}", handler.Update).Methods("PUT")
			protected.HandleFunc("/{shop_id}/orders/{order_id:[0-9]+}", handler.GetByID).Methods("GET")
			protected.HandleFunc("/{shop_id}/orders", handler.GetAll).Methods("GET")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// SetupMetricsTestApp initializes the test application for metrics tests (protected endpoints)
func (ctx *TestContext) SetupMetricsTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Allow queries to be executed in any order (dashboard runs 7 queries in parallel)
	sqlMock.MatchExpectationsInOrder(false)

	// Create TokenService for auth middleware
	tokenService := jwt.NewTokenService()

	// Generate test token for authenticated requests
	// - Shop 1, 2: standard test shops
	// - Shop 888: for "not found" tests (passes auth, but no data)
	// - Shop 999: intentionally excluded for "not owned" tests
	testUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	testShopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}, {ShopID: 2, Role: "admin"}, {ShopID: 888, Role: "owner"}}
	ctx.authToken, err = tokenService.Generate(context.Background(), testUser, testShopRoles)
	if err != nil {
		return err
	}

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide TokenService for auth middleware
			fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),

			// Provide auth middleware
			middleware.NewAuthMiddleware,

			// Provide metrics dependencies
			fx.Annotate(postgresql.NewMetricsRepository, fx.As(new(ports.MetricsRepository))),
			fx.Annotate(services.NewMetricsService, fx.As(new(ports.MetricsService))),

			// Provide use cases
			fx.Annotate(metrics.NewGetDashboardUseCase, fx.As(new(ports.GetDashboardMetricsUseCase))),
			fx.Annotate(metrics.NewGetRevenueTrendUseCase, fx.As(new(ports.GetRevenueTrendUseCase))),
			fx.Annotate(metrics.NewGetTopProductsUseCase, fx.As(new(ports.GetTopProductsUseCase))),
			fx.Annotate(metrics.NewGetTopCustomersUseCase, fx.As(new(ports.GetTopCustomersUseCase))),
			fx.Annotate(metrics.NewGetShippingSummaryUseCase, fx.As(new(ports.GetShippingSummaryUseCase))),
			fx.Annotate(metrics.NewGetVisitsTrendUseCase, fx.As(new(ports.GetVisitsTrendUseCase))),

			// Provide handler
			fx.Annotate(authhttp.NewMetricsHandler, fx.As(new(ports.MetricsHandler))),

			// Provide shop ownership middleware
			middleware.NewShopOwnershipMiddleware,
		),
		fx.Invoke(func(handler ports.MetricsHandler, authMiddleware *middleware.AuthMiddleware, shopOwnershipMiddleware *middleware.ShopOwnershipMiddleware) {
			// Create HTTP router and server
			router := mux.NewRouter()

			// Protected routes (require auth + shop ownership)
			protected := router.PathPrefix("/shops").Subrouter()
			protected.Use(authMiddleware.Authenticate)
			protected.Use(shopOwnershipMiddleware.Authorize)
			protected.HandleFunc("/{shop_id}/metrics/dashboard", handler.GetDashboard).Methods("GET")
			protected.HandleFunc("/{shop_id}/metrics/revenue/trend", handler.GetRevenueTrend).Methods("GET")
			protected.HandleFunc("/{shop_id}/metrics/products/top", handler.GetTopProducts).Methods("GET")
			protected.HandleFunc("/{shop_id}/metrics/customers/top", handler.GetTopCustomers).Methods("GET")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// TeardownTestApp cleans up the test application
func (ctx *TestContext) TeardownTestApp() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.app != nil {
		err := ctx.app.Stop(context.Background())
		if err != nil {
			return err
		}
	}
	if ctx.mockDB != nil {
		ctx.mockDB.Close()
	}
	if ctx.server != nil {
		ctx.server.Close()
	}
	return nil
}
