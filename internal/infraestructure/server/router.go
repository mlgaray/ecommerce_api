package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	httpAdapter "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/middleware"
)

type Router interface {
	RouteApp() *mux.Router
}
type router struct {
	router                  *mux.Router
	authHandler             ports.AuthHandler
	healthHandler           ports.HealthHandler
	productHandler          ports.ProductHandler
	categoryHandler         ports.CategoryHandler
	shopHandler             ports.ShopHandler
	storeHandler            ports.StoreHandler
	orderHandler            ports.OrderHandler
	couponHandler           ports.CouponHandler
	staffHandler            ports.StaffHandler
	metricsHandler          ports.MetricsHandler
	orderWSHandler          *httpAdapter.OrderWSHandler
	authMiddleware          *middleware.AuthMiddleware
	shopOwnershipMiddleware *middleware.ShopOwnershipMiddleware
	permissionMiddleware    *middleware.PermissionMiddleware
}

func NewRouter(
	authHandler ports.AuthHandler,
	healthHandler ports.HealthHandler,
	productHandler ports.ProductHandler,
	categoryHandler ports.CategoryHandler,
	shopHandler ports.ShopHandler,
	storeHandler ports.StoreHandler,
	orderHandler ports.OrderHandler,
	couponHandler ports.CouponHandler,
	staffHandler ports.StaffHandler,
	metricsHandler ports.MetricsHandler,
	orderWSHandler *httpAdapter.OrderWSHandler,
	authMiddleware *middleware.AuthMiddleware,
	shopOwnershipMiddleware *middleware.ShopOwnershipMiddleware,
	permissionMiddleware *middleware.PermissionMiddleware,
) *router {
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.PrometheusMiddleware)
	return &router{
		router:                  r,
		authHandler:             authHandler,
		healthHandler:           healthHandler,
		productHandler:          productHandler,
		categoryHandler:         categoryHandler,
		shopHandler:             shopHandler,
		storeHandler:            storeHandler,
		orderHandler:            orderHandler,
		couponHandler:           couponHandler,
		staffHandler:            staffHandler,
		metricsHandler:          metricsHandler,
		orderWSHandler:          orderWSHandler,
		authMiddleware:          authMiddleware,
		shopOwnershipMiddleware: shopOwnershipMiddleware,
		permissionMiddleware:    permissionMiddleware,
	}
}

func (r *router) RouteApp() *mux.Router {
	r.healthRoutes()
	r.authRoutes()
	r.productRoutes()
	r.categoryRoutes()
	r.metricsRoutes()
	r.shopMetricsRoutes()
	r.couponRoutes()
	r.staffRoutes()
	r.shopRoutes()
	r.storeRoutes()
	r.websocketRoutes()
	return r.router
}

func (r *router) healthRoutes() {
	r.router.HandleFunc("/health", r.healthHandler.Health).Methods(http.MethodGet)
}

func (r *router) authRoutes() {
	sub := r.router.PathPrefix("/auth").Subrouter()
	sub.HandleFunc("/signin", r.authHandler.SignIn).Methods(http.MethodPost)
	sub.HandleFunc("/signup", r.authHandler.SignUp).Methods(http.MethodPost)
}

func (r *router) productRoutes() {
	perm := r.permissionMiddleware
	sub := r.router.PathPrefix("/products").Subrouter()
	sub.Use(r.authMiddleware.Authenticate)
	sub.Handle("", perm.RequirePermission("create_product")(http.HandlerFunc(r.productHandler.Create))).Methods(http.MethodPost)
	sub.HandleFunc("/{product_id}", r.productHandler.GetByID).Methods(http.MethodGet)
	sub.Handle("/{product_id}", perm.RequirePermission("update_product")(http.HandlerFunc(r.productHandler.Update))).Methods(http.MethodPut)
	sub.Handle("/{product_id}", perm.RequirePermission("delete_product")(http.HandlerFunc(r.productHandler.Delete))).Methods(http.MethodDelete)
}

func (r *router) categoryRoutes() {
	perm := r.permissionMiddleware
	sub := r.router.PathPrefix("/categories").Subrouter()
	sub.Use(r.authMiddleware.Authenticate)
	sub.Handle("", perm.RequirePermission("create_category")(http.HandlerFunc(r.categoryHandler.Create))).Methods(http.MethodPost)
	sub.HandleFunc("/{category_id}", r.categoryHandler.GetByID).Methods(http.MethodGet)
	sub.Handle("/{category_id}", perm.RequirePermission("update_category")(http.HandlerFunc(r.categoryHandler.Update))).Methods(http.MethodPut)
	sub.Handle("/{category_id}", perm.RequirePermission("delete_category")(http.HandlerFunc(r.categoryHandler.Delete))).Methods(http.MethodDelete)
}

func (r *router) couponRoutes() {
	perm := r.permissionMiddleware
	protected := r.router.PathPrefix("/shops").Subrouter()
	protected.Use(r.authMiddleware.Authenticate)
	protected.Use(r.shopOwnershipMiddleware.Authorize)
	protected.Handle("/{shop_id}/coupons", perm.RequirePermission("create_coupon")(http.HandlerFunc(r.couponHandler.Create))).Methods(http.MethodPost)
	protected.Handle("/{shop_id}/coupons", perm.RequirePermission("view_coupons")(http.HandlerFunc(r.couponHandler.GetAll))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/coupons/{coupon_id:[0-9]+}", perm.RequirePermission("view_coupons")(http.HandlerFunc(r.couponHandler.GetByID))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/coupons/{coupon_id:[0-9]+}", perm.RequirePermission("update_coupon")(http.HandlerFunc(r.couponHandler.Update))).Methods(http.MethodPut)
	protected.Handle("/{shop_id}/coupons/{coupon_id:[0-9]+}", perm.RequirePermission("delete_coupon")(http.HandlerFunc(r.couponHandler.Delete))).Methods(http.MethodDelete)
}

func (r *router) staffRoutes() {
	perm := r.permissionMiddleware
	protected := r.router.PathPrefix("/shops").Subrouter()
	protected.Use(r.authMiddleware.Authenticate)
	protected.Use(r.shopOwnershipMiddleware.Authorize)
	protected.Handle("/{shop_id}/staff", perm.RequirePermission("create_staff")(http.HandlerFunc(r.staffHandler.Create))).Methods(http.MethodPost)
	protected.Handle("/{shop_id}/staff", perm.RequirePermission("view_staff")(http.HandlerFunc(r.staffHandler.GetAll))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/staff/{staff_id:[0-9]+}", perm.RequirePermission("view_staff")(http.HandlerFunc(r.staffHandler.GetByID))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/staff/{staff_id:[0-9]+}", perm.RequirePermission("update_staff")(http.HandlerFunc(r.staffHandler.Update))).Methods(http.MethodPut)
	protected.Handle("/{shop_id}/staff/{staff_id:[0-9]+}", perm.RequirePermission("delete_staff")(http.HandlerFunc(r.staffHandler.Delete))).Methods(http.MethodDelete)
	protected.Handle("/{shop_id}/staff/{staff_id:[0-9]+}/status", perm.RequirePermission("update_staff")(http.HandlerFunc(r.staffHandler.ToggleStatus))).Methods(http.MethodPatch)
}

func (r *router) shopRoutes() {
	perm := r.permissionMiddleware
	// Public routes (no auth required)
	sub := r.router.PathPrefix("/shops").Subrouter()
	sub.HandleFunc("/{shop_id}/products", r.productHandler.GetAllByShopIDWithFilters).Methods(http.MethodGet)
	sub.HandleFunc("/{shop_id}/categories", r.categoryHandler.GetAllByShopIDWithFilters).Methods(http.MethodGet)

	// Protected routes (auth + shop ownership + permissions)
	protected := r.router.PathPrefix("/shops").Subrouter()
	protected.Use(r.authMiddleware.Authenticate)
	protected.Use(r.shopOwnershipMiddleware.Authorize)
	// Orders — view
	protected.Handle("/{shop_id}/orders/{order_id:[0-9]+}", perm.RequirePermission("view_orders")(http.HandlerFunc(r.orderHandler.GetByID))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/orders", perm.RequirePermission("view_orders")(http.HandlerFunc(r.orderHandler.GetAll))).Methods(http.MethodGet)
	// Orders — manage
	protected.Handle("/{shop_id}/orders/{order_id:[0-9]+}/status", perm.RequirePermission("manage_orders")(http.HandlerFunc(r.orderHandler.UpdateStatus))).Methods(http.MethodPatch)
	protected.Handle("/{shop_id}/orders/{order_id:[0-9]+}/coupon", perm.RequirePermission("manage_orders")(http.HandlerFunc(r.orderHandler.RemoveOrderCoupon))).Methods(http.MethodDelete)
	protected.Handle("/{shop_id}/orders/{order_id:[0-9]+}", perm.RequirePermission("manage_orders")(http.HandlerFunc(r.orderHandler.Update))).Methods(http.MethodPut)
	// Shop settings
	protected.Handle("/{shop_id}", perm.RequirePermission("update_shop")(http.HandlerFunc(r.shopHandler.GetByID))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}", perm.RequirePermission("update_shop")(http.HandlerFunc(r.shopHandler.Update))).Methods(http.MethodPut)
}

func (r *router) metricsRoutes() {
	r.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
}

func (r *router) shopMetricsRoutes() {
	perm := r.permissionMiddleware
	protected := r.router.PathPrefix("/shops").Subrouter()
	protected.Use(r.authMiddleware.Authenticate)
	protected.Use(r.shopOwnershipMiddleware.Authorize)
	protected.Handle("/{shop_id}/metrics/dashboard", perm.RequirePermission("view_metrics")(http.HandlerFunc(r.metricsHandler.GetDashboard))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/metrics/revenue/trend", perm.RequirePermission("view_metrics")(http.HandlerFunc(r.metricsHandler.GetRevenueTrend))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/metrics/products/top", perm.RequirePermission("view_metrics")(http.HandlerFunc(r.metricsHandler.GetTopProducts))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/metrics/customers/top", perm.RequirePermission("view_metrics")(http.HandlerFunc(r.metricsHandler.GetTopCustomers))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/metrics/shipping/summary", perm.RequirePermission("view_metrics")(http.HandlerFunc(r.metricsHandler.GetShippingSummary))).Methods(http.MethodGet)
	protected.Handle("/{shop_id}/metrics/visits/trend", perm.RequirePermission("view_metrics")(http.HandlerFunc(r.metricsHandler.GetVisitsTrend))).Methods(http.MethodGet)
}

func (r *router) storeRoutes() {
	// Public routes (no auth required) - Customer-facing store endpoints
	sub := r.router.PathPrefix("/stores").Subrouter()
	sub.HandleFunc("/slugs/{slug}/check-availability", r.storeHandler.CheckSlugAvailability).Methods(http.MethodGet)
	sub.HandleFunc("/{slug}/products/featured", r.storeHandler.GetFeaturedProducts).Methods(http.MethodGet)
	sub.HandleFunc("/{slug}/products/{productId}", r.storeHandler.GetProductByID).Methods(http.MethodGet)
	sub.HandleFunc("/{slug}/products", r.storeHandler.GetProducts).Methods(http.MethodGet)
	sub.HandleFunc("/{slug}/categories", r.storeHandler.GetCategories).Methods(http.MethodGet)
	sub.HandleFunc("/{slug}/coupons/validate", r.storeHandler.ValidateCoupon).Methods(http.MethodPost)
	sub.HandleFunc("/{slug}/orders", r.orderHandler.Create).Methods(http.MethodPost)
	sub.HandleFunc("/{slug}", r.storeHandler.GetBySlug).Methods(http.MethodGet)
}

func (r *router) websocketRoutes() {
	r.router.HandleFunc("/shops/{shop_id}/orders/ws", r.orderWSHandler.Connect).Methods(http.MethodGet)
}
