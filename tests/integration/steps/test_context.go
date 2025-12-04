package steps

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/auth"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/category"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/product"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/jwt"
	authhttp "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http"
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

	// Test control
	scenario string

	// SQL Mock
	mockDB      *sql.DB
	mockSQLMock sqlmock.Sqlmock
}

// Global test context instance
var testCtx *TestContext

// GetTestContext returns the current test context
func GetTestContext() *TestContext {
	if testCtx == nil {
		testCtx = &TestContext{}
	}
	return testCtx
}

// Reset clears all test context data
func (ctx *TestContext) Reset() {
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
	// Initialize logger for tests
	logs.Init()

	// Setup SQL mock
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		return err
	}
	ctx.mockDB = db
	ctx.mockSQLMock = sqlMock

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide real services with interface annotations
			fx.Annotate(services.NewUserService, fx.As(new(ports.UserService))),
			fx.Annotate(services.NewAuthService, fx.As(new(ports.AuthService))),
			fx.Annotate(services.NewSignupService, fx.As(new(ports.SignUpService))),
			fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),
			fx.Annotate(postgresql.NewUserRepository, fx.As(new(ports.UserRepository))),
			fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),
			fx.Annotate(postgresql.NewRoleRepository, fx.As(new(ports.RoleRepository))),
			fx.Annotate(postgresql.NewSignupRepository, fx.As(new(ports.SignupRepository))),

			// Provide use cases
			auth.NewSignInUseCase,
			auth.NewSignUpUseCase,

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

// SetupProductTestApp initializes the test application for product tests
func (ctx *TestContext) SetupProductTestApp() error {
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

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

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

			// Provide handler
			authhttp.NewProductHandler,
		),
		fx.Invoke(func(handler *authhttp.ProductHandler) {
			// Create HTTP router and server
			router := mux.NewRouter()
			router.HandleFunc("/products", handler.Create).Methods("POST")
			router.HandleFunc("/shops/{shop_id}/products", handler.GetAllByShopIDWithFilters).Methods("GET")
			router.HandleFunc("/products/{product_id}", handler.GetByID).Methods("GET")
			router.HandleFunc("/products/{product_id}", handler.Update).Methods("PUT")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// SetupCategoryTestApp initializes the test application for category tests
func (ctx *TestContext) SetupCategoryTestApp() error {
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

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

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
			fx.Annotate(category.NewGetAllByShopIDWithFiltersUseCase, fx.As(new(ports.GetAllCategoriesByShopIDWithFiltersUseCase))),

			// Provide handler
			authhttp.NewCategoryHandler,
		),
		fx.Invoke(func(handler *authhttp.CategoryHandler) {
			// Create HTTP router and server
			router := mux.NewRouter()
			router.HandleFunc("/categories", handler.Create).Methods("POST")
			router.HandleFunc("/shops/{shop_id}/categories", handler.GetAllByShopIDWithFilters).Methods("GET")

			ctx.server = httptest.NewServer(router)
		}),
		fx.NopLogger, // Suppress fx logs during tests
	)

	return ctx.app.Start(context.Background())
}

// TeardownTestApp cleans up the test application
func (ctx *TestContext) TeardownTestApp() error {
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
