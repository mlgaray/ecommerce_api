package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/auth"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/category"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/product"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/shop"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/store"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/assets"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/jwt"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/middleware"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/repositories/postgresql"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/server"
)

var Module = fx.Options(
	fx.Provide(
		// TOKEN
		fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),

		// AUTH MIDDLEWARE
		middleware.NewAuthMiddleware,

		// AUTH
		fx.Annotate(http.NewAuthHandler, fx.As(new(ports.AuthHandler))),
		fx.Annotate(services.NewAuthService, fx.As(new(ports.AuthService))),

		// HEALTH
		fx.Annotate(http.NewHealthHandler, fx.As(new(ports.HealthHandler))),

		// USER
		fx.Annotate(services.NewUserService, fx.As(new(ports.UserService))),
		fx.Annotate(postgresql.NewUserRepository, fx.As(new(ports.UserRepository))),

		// ROLE
		fx.Annotate(postgresql.NewRoleRepository, fx.As(new(ports.RoleRepository))),

		// PAYMENT METHODS (needed by ShopRepository)
		fx.Annotate(postgresql.NewPaymentMethodRepository, fx.As(new(ports.PaymentMethodRepository))),

		// DELIVERY METHODS (needed by ShopRepository)
		fx.Annotate(postgresql.NewDeliveryMethodRepository, fx.As(new(ports.DeliveryMethodRepository))),

		// SHOP (depends on PaymentMethodRepository and DeliveryMethodRepository)
		fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),

		// Sign UP
		fx.Annotate(services.NewSignupService, fx.As(new(ports.SignUpService))),
		fx.Annotate(postgresql.NewSignupRepository, fx.As(new(ports.SignupRepository))),

		fx.Annotate(auth.NewSignInUseCase, fx.As(new(ports.SignInUseCase))),
		fx.Annotate(auth.NewSignUpUseCase, fx.As(new(ports.SignUpUseCase))),

		// PAGINATION (shared services for cursor-based pagination)
		fx.Annotate(services.NewPaginationService[*models.Product], fx.As(new(ports.PaginationService[*models.Product]))),
		fx.Annotate(services.NewPaginationService[*models.Category], fx.As(new(ports.PaginationService[*models.Category]))),

		// PRODUCT
		// Repository first (no dependencies)
		fx.Annotate(postgresql.NewProductRepository, fx.As(new(ports.ProductRepository))),
		// Service depends on Repository
		fx.Annotate(services.NewProductService, fx.As(new(ports.ProductService))),
		// Use Cases depend on Services (NOT repositories)
		fx.Annotate(product.NewCreateProductUseCase, fx.As(new(ports.CreateProductUseCase))),
		fx.Annotate(product.NewGetAllByShopIDWithFiltersUseCase, fx.As(new(ports.GetAllByShopIDWithFiltersUseCase))),
		fx.Annotate(product.NewGetByIDUseCase, fx.As(new(ports.GetByIDUseCase))),
		fx.Annotate(product.NewUpdateProductUseCase, fx.As(new(ports.UpdateProductUseCase))),
		fx.Annotate(product.NewDeleteProductUseCase, fx.As(new(ports.DeleteProductUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewProductHandler, fx.As(new(ports.ProductHandler))),

		// ASSETS (Cloudinary)
		fx.Annotate(assets.NewCloudinaryConnection, fx.As(new(assets.CloudinaryConnection))),
		fx.Annotate(assets.NewCloudinaryAssetService, fx.As(new(ports.AssetService))),

		// CATEGORY
		// Repository first (no dependencies)
		fx.Annotate(postgresql.NewCategoryRepository, fx.As(new(ports.CategoryRepository))),
		// Service depends on Repository + AssetService
		fx.Annotate(services.NewCategoryService, fx.As(new(ports.CategoryService))),
		// Use Cases depend on Services
		fx.Annotate(category.NewCreateCategoryUseCase, fx.As(new(ports.CreateCategoryUseCase))),
		fx.Annotate(category.NewUpdateCategoryUseCase, fx.As(new(ports.UpdateCategoryUseCase))),
		fx.Annotate(category.NewDeleteCategoryUseCase, fx.As(new(ports.DeleteCategoryUseCase))),
		fx.Annotate(category.NewGetByIDUseCase, fx.As(new(ports.GetCategoryByIDUseCase))),
		fx.Annotate(category.NewGetAllByShopIDWithFiltersUseCase, fx.As(new(ports.GetAllCategoriesByShopIDWithFiltersUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewCategoryHandler, fx.As(new(ports.CategoryHandler))),

		// SHOP
		// Service depends on Repository + AssetService
		fx.Annotate(services.NewShopService, fx.As(new(ports.ShopService))),
		// Use Cases depend on Services (NOT repositories)
		fx.Annotate(shop.NewGetShopByIDUseCase, fx.As(new(ports.GetShopByIDUseCase))),
		fx.Annotate(shop.NewUpdateShopUseCase, fx.As(new(ports.UpdateShopUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewShopHandler, fx.As(new(ports.ShopHandler))),

		// STORE (public storefront endpoints)
		// Service depends on ShopRepository (reuses existing repository)
		fx.Annotate(services.NewStoreService, fx.As(new(ports.StoreService))),
		// Use Cases depend on Services
		fx.Annotate(store.NewGetStoreBySlugUseCase, fx.As(new(ports.GetStoreBySlugUseCase))),
		fx.Annotate(store.NewGetStoreCategoriesUseCase, fx.As(new(ports.GetStoreCategoriesUseCase))),
		fx.Annotate(store.NewGetStoreProductsUseCase, fx.As(new(ports.GetStoreProductsUseCase))),
		fx.Annotate(store.NewGetStoreFeaturedProductsUseCase, fx.As(new(ports.GetStoreFeaturedProductsUseCase))),
		fx.Annotate(store.NewGetStoreProductByIDUseCase, fx.As(new(ports.GetStoreProductByIDUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewStoreHandler, fx.As(new(ports.StoreHandler))),

		// SERVER
		server.NewServer,
		fx.Annotate(server.NewRouter, fx.As(new(server.Router))),

		fx.Annotate(postgresql.NewDataBaseConnection, fx.As(new(postgresql.DataBaseConnection))),
	),
	fx.Invoke(
		RegisterHooks,
		InitializeLogger,
	),
)

func main() {
	log.Println("Starting application...")
	app := fx.New(Module, fx.StartTimeout(30*time.Second))
	app.Run()
	if err := app.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	// Manejador de señales del sistema
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Escuchar las señales del sistema en una goroutine
	go func() {
		<-signals
		// Detener la aplicación cuando se recibe una señal del sistema
		if err := app.Stop(context.Background()); err != nil {
			log.Fatalf("Failed to stop: %v", err)
		}
	}()
}

func InitializeLogger() {
	logs.Init()
}

func RegisterHooks(lc fx.Lifecycle, server *server.Server) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			server.Initialize()
			return nil
		},
		OnStop: func(context.Context) error {
			return nil
		},
	})
}

// func NewServerHooks(router *mux.Router) fx.Hook {
//	return fx.Hook{
//		OnStart: func(context.Context) error {
//			handler := cors.AllowAll().Handler(router)
//			log.Fatal(http.ListenAndServe(":"+"8080", handler))
//			/*if err != nil {
//				return fmt.Errorf("failed to initialize server: %w", err)
//			}*/
//			return nil
//		},
//		OnStop: func(context.Context) error {
//			// return server.Stop()
//			return nil
//		},
//	}
//}

// var totalRequests = prometheus.NewCounterVec(
//	prometheus.CounterOpts{
//		Name: "http_requests_total",
//		Help: "Number of get requests.",
//	},
//	[]string{"path"},
//)

// func prometheusMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		rw := NewResponseWriter(w)
//		next.ServeHTTP(rw, r)
//
//		totalRequests.WithLabelValues(path).Inc()
//	})
//}
