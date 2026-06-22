package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"go.uber.org/fx"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/auth"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/category"
	couponUC "github.com/mlgaray/ecommerce_api/internal/application/usecases/coupon"
	metricsUC "github.com/mlgaray/ecommerce_api/internal/application/usecases/metrics"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/order"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/product"
	roleUC "github.com/mlgaray/ecommerce_api/internal/application/usecases/role"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/shop"
	staffUC "github.com/mlgaray/ecommerce_api/internal/application/usecases/staff"
	"github.com/mlgaray/ecommerce_api/internal/application/usecases/store"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/assets"
	authBcrypt "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/bcrypt"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/jwt"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/middleware"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	appMetrics "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/metrics"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/repositories/postgresql"
	websocketAdapter "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/websocket"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/server"
)

var Module = fx.Options(
	fx.Provide(
		// TOKEN
		fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),

		// AUTH MIDDLEWARE
		middleware.NewAuthMiddleware,

		// SHOP OWNERSHIP MIDDLEWARE
		middleware.NewShopOwnershipMiddleware,

		// AUTH
		fx.Annotate(http.NewAuthHandler, fx.As(new(ports.AuthHandler))),
		fx.Annotate(authBcrypt.NewAuthService, fx.As(new(ports.AuthService))),

		// DATABASE — singleton *sql.DB exposed via DataBaseConnection.Connect()
		fx.Annotate(postgresql.NewDataBaseConnection, fx.As(new(postgresql.DataBaseConnection))),
		func(conn postgresql.DataBaseConnection) *sql.DB { return conn.Connect() },

		// HEALTH — depends on *sql.DB for DB ping
		fx.Annotate(http.NewHealthHandler, fx.As(new(ports.HealthHandler))),

		// USER
		fx.Annotate(services.NewUserService, fx.As(new(ports.UserService))),
		fx.Annotate(postgresql.NewUserRepository, fx.As(new(ports.UserRepository))),

		// ROLE
		fx.Annotate(postgresql.NewRoleRepository, fx.As(new(ports.RoleRepository))),
		fx.Annotate(services.NewRoleService, fx.As(new(ports.RoleService))),
		fx.Annotate(roleUC.NewGetAssignableRolesUseCase, fx.As(new(ports.GetAssignableRolesUseCase))),
		fx.Annotate(http.NewRoleHandler, fx.As(new(ports.RoleHandler))),

		// PAYMENT METHODS (needed by ShopRepository)
		fx.Annotate(postgresql.NewPaymentMethodRepository, fx.As(new(ports.PaymentMethodRepository))),

		// DELIVERY METHODS (needed by ShopRepository)
		fx.Annotate(postgresql.NewDeliveryMethodRepository, fx.As(new(ports.DeliveryMethodRepository))),

		// SHOP (depends on PaymentMethodRepository and DeliveryMethodRepository)
		fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),

		// PERMISSIONS (in-memory role→permissions map)
		fx.Annotate(postgresql.NewPermissionRepository, fx.As(new(ports.PermissionRepository))),
		services.NewPermissionService,
		func(svc *services.PermissionServiceImpl) ports.PermissionService { return svc },

		// PERMISSION MIDDLEWARE
		middleware.NewPermissionMiddleware,

		// STAFF
		fx.Annotate(postgresql.NewStaffRepository, fx.As(new(ports.StaffRepository))),
		fx.Annotate(services.NewStaffService, fx.As(new(ports.StaffService))),
		fx.Annotate(staffUC.NewCreateStaffUseCase, fx.As(new(ports.CreateStaffUseCase))),
		fx.Annotate(staffUC.NewGetAllStaffByShopIDUseCase, fx.As(new(ports.GetAllStaffByShopIDUseCase))),
		fx.Annotate(staffUC.NewGetStaffByIDUseCase, fx.As(new(ports.GetStaffByIDUseCase))),
		fx.Annotate(staffUC.NewUpdateStaffUseCase, fx.As(new(ports.UpdateStaffUseCase))),
		fx.Annotate(staffUC.NewDeleteStaffUseCase, fx.As(new(ports.DeleteStaffUseCase))),
		fx.Annotate(staffUC.NewToggleStaffStatusUseCase, fx.As(new(ports.ToggleStaffStatusUseCase))),
		fx.Annotate(http.NewStaffHandler, fx.As(new(ports.StaffHandler))),

		// Sign UP
		fx.Annotate(services.NewSignupService, fx.As(new(ports.SignUpService))),
		fx.Annotate(postgresql.NewSignupRepository, fx.As(new(ports.SignupRepository))),

		fx.Annotate(auth.NewSignInUseCase, fx.As(new(ports.SignInUseCase))),
		fx.Annotate(auth.NewSignUpUseCase, fx.As(new(ports.SignUpUseCase))),

		// PAGINATION (shared services for cursor-based pagination)
		fx.Annotate(services.NewPaginationService[*models.Product], fx.As(new(ports.PaginationService[*models.Product]))),
		fx.Annotate(services.NewPaginationService[*models.Category], fx.As(new(ports.PaginationService[*models.Category]))),
		fx.Annotate(services.NewPaginationService[*models.Order], fx.As(new(ports.PaginationService[*models.Order]))),
		fx.Annotate(services.NewPaginationService[*models.Coupon], fx.As(new(ports.PaginationService[*models.Coupon]))),
		fx.Annotate(services.NewPaginationService[*models.Staff], fx.As(new(ports.PaginationService[*models.Staff]))),

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

		// STORE VISIT
		fx.Annotate(postgresql.NewStoreVisitRepository, fx.As(new(ports.StoreVisitRepository))),

		// STORE (public storefront endpoints)
		// Service depends on ShopRepository, ProductRepository, StoreVisitRepository
		fx.Annotate(services.NewStoreService, fx.As(new(ports.StoreService))),
		// Use Cases depend on Services
		fx.Annotate(store.NewCheckSlugAvailabilityUseCase, fx.As(new(ports.CheckSlugAvailabilityUseCase))),
		fx.Annotate(store.NewGetStoreBySlugUseCase, fx.As(new(ports.GetStoreBySlugUseCase))),
		fx.Annotate(store.NewGetStoreCategoriesUseCase, fx.As(new(ports.GetStoreCategoriesUseCase))),
		fx.Annotate(store.NewGetStoreProductsUseCase, fx.As(new(ports.GetStoreProductsUseCase))),
		fx.Annotate(store.NewGetStoreFeaturedProductsUseCase, fx.As(new(ports.GetStoreFeaturedProductsUseCase))),
		fx.Annotate(store.NewGetStoreProductByIDUseCase, fx.As(new(ports.GetStoreProductByIDUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewStoreHandler, fx.As(new(ports.StoreHandler))),

		// WEBSOCKET
		// Hub manages WebSocket connections (concrete type for OrderWSHandler)
		websocketAdapter.NewHub,
		// OrderEventNotifier port (Hub implements this interface)
		func(hub *websocketAdapter.Hub) ports.OrderEventNotifier { return hub },
		// WebSocket handler for order notifications
		http.NewOrderWSHandler,

		// COUPON
		// Repository first (no dependencies)
		fx.Annotate(postgresql.NewCouponRepository, fx.As(new(ports.CouponRepository))),
		// Service depends on Repository
		fx.Annotate(services.NewCouponService, fx.As(new(ports.CouponService))),
		// Use Cases depend on Services
		fx.Annotate(couponUC.NewCreateCouponUseCase, fx.As(new(ports.CreateCouponUseCase))),
		fx.Annotate(couponUC.NewUpdateCouponUseCase, fx.As(new(ports.UpdateCouponUseCase))),
		fx.Annotate(couponUC.NewDeleteCouponUseCase, fx.As(new(ports.DeleteCouponUseCase))),
		fx.Annotate(couponUC.NewGetCouponByIDUseCase, fx.As(new(ports.GetCouponByIDUseCase))),
		fx.Annotate(couponUC.NewGetAllCouponsByShopIDUseCase, fx.As(new(ports.GetAllCouponsByShopIDUseCase))),
		fx.Annotate(couponUC.NewValidateStoreCouponUseCase, fx.As(new(ports.ValidateStoreCouponUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewCouponHandler, fx.As(new(ports.CouponHandler))),

		// ORDER
		// Repository first (no dependencies)
		fx.Annotate(postgresql.NewOrderRepository, fx.As(new(ports.OrderRepository))),
		// Service depends on Repositories
		fx.Annotate(services.NewOrderService, fx.As(new(ports.OrderService))),
		// Use Cases depend on Services
		fx.Annotate(order.NewCreateOrderUseCase, fx.As(new(ports.CreateOrderUseCase))),
		fx.Annotate(order.NewGetAllOrdersByShopIDUseCase, fx.As(new(ports.GetAllOrdersByShopIDUseCase))),
		fx.Annotate(order.NewGetOrderByIDUseCase, fx.As(new(ports.GetOrderByIDUseCase))),
		fx.Annotate(order.NewUpdateOrderStatusUseCase, fx.As(new(ports.UpdateOrderStatusUseCase))),
		fx.Annotate(order.NewUpdateOrderUseCase, fx.As(new(ports.UpdateOrderUseCase))),
		fx.Annotate(order.NewRemoveOrderCouponUseCase, fx.As(new(ports.RemoveOrderCouponUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewOrderHandler, fx.As(new(ports.OrderHandler))),

		// METRICS
		// Repository first (no dependencies)
		fx.Annotate(postgresql.NewMetricsRepository, fx.As(new(ports.MetricsRepository))),
		// Service depends on Repository
		fx.Annotate(services.NewMetricsService, fx.As(new(ports.MetricsService))),
		// Use Cases depend on MetricsService only (timezone comes from frontend)
		fx.Annotate(metricsUC.NewGetDashboardUseCase, fx.As(new(ports.GetDashboardMetricsUseCase))),
		fx.Annotate(metricsUC.NewGetRevenueTrendUseCase, fx.As(new(ports.GetRevenueTrendUseCase))),
		fx.Annotate(metricsUC.NewGetTopProductsUseCase, fx.As(new(ports.GetTopProductsUseCase))),
		fx.Annotate(metricsUC.NewGetTopCustomersUseCase, fx.As(new(ports.GetTopCustomersUseCase))),
		fx.Annotate(metricsUC.NewGetShippingSummaryUseCase, fx.As(new(ports.GetShippingSummaryUseCase))),
		fx.Annotate(metricsUC.NewGetVisitsTrendUseCase, fx.As(new(ports.GetVisitsTrendUseCase))),
		// Handler depends on Use Cases
		fx.Annotate(http.NewMetricsHandler, fx.As(new(ports.MetricsHandler))),

		// SERVER
		server.NewServer,
		fx.Annotate(server.NewRouter, fx.As(new(server.Router))),
	),
	fx.Invoke(
		RegisterHooks,
		InitializeLogger,
		// Register Go runtime, process, and DB pool metric collectors
		appMetrics.RegisterRuntimeCollectors,
	),
)

func main() {
	log.Println("Starting application...")
	fx.New(Module, fx.StartTimeout(30*time.Second)).Run()
}

func InitializeLogger() {
	logs.Init()
}

func RegisterHooks(lc fx.Lifecycle, server *server.Server, hub *websocketAdapter.Hub, permissionService *services.PermissionServiceImpl) {
	var cancelPermissionRefresh context.CancelFunc

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			server.Initialize()

			// Start permission refresh loop (reloads role→permissions every 5 min)
			refreshCtx, cancel := context.WithCancel(context.Background())
			cancelPermissionRefresh = cancel
			go permissionService.StartRefreshLoop(refreshCtx, 5*time.Minute)

			return nil
		},
		OnStop: func(context.Context) error {
			if cancelPermissionRefresh != nil {
				cancelPermissionRefresh()
			}
			hub.CloseAll()
			return nil
		},
	})
}
