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
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/product"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/jwt"
	authhttp "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/repositories/postgresql"
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

	// Requests/Responses
	signInRequest  contracts.SignInRequest
	signInResponse contracts.SignInResponse
	signUpRequest  contracts.SignUpRequest
	successMessage string
	errorMessage   string

	// Product requests/responses
	productRequest   models.Product
	productImages    [][]byte
	productShopID    int
	createdProduct   *models.Product
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
	ctx.signInRequest = contracts.SignInRequest{}
	ctx.signInResponse = contracts.SignInResponse{}
	ctx.signUpRequest = contracts.SignUpRequest{}
	ctx.successMessage = ""
	ctx.errorMessage = ""
	ctx.productRequest = models.Product{}
	ctx.productImages = nil
	ctx.productShopID = 0
	ctx.createdProduct = nil
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

	// Create FX app with real services but mocked DB
	ctx.app = fx.New(
		fx.Provide(
			// Provide mocked database connection
			func() postgresql.DataBaseConnection {
				return &mockDataBaseConnection{db: db}
			},

			// Provide product dependencies
			fx.Annotate(services.NewProductService, fx.As(new(ports.ProductService))),
			fx.Annotate(postgresql.NewProductRepository, fx.As(new(ports.ProductRepository))),
			fx.Annotate(product.NewCreateProductUseCase, fx.As(new(ports.CreateProductUseCase))),

			// Provide handler
			authhttp.NewProductHandler,
		),
		fx.Invoke(func(handler *authhttp.ProductHandler) {
			// Create HTTP router and server
			router := mux.NewRouter()
			router.HandleFunc("/products", handler.Create).Methods("POST")

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
