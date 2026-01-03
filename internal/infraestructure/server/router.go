package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/middleware"
)

type Router interface {
	RouteApp() *mux.Router
}
type router struct {
	router          *mux.Router
	authHandler     ports.AuthHandler
	healthHandler   ports.HealthHandler
	productHandler  ports.ProductHandler
	categoryHandler ports.CategoryHandler
	shopHandler     ports.ShopHandler
	authMiddleware  *middleware.AuthMiddleware
}

func NewRouter(
	authHandler ports.AuthHandler,
	healthHandler ports.HealthHandler,
	productHandler ports.ProductHandler,
	categoryHandler ports.CategoryHandler,
	shopHandler ports.ShopHandler,
	authMiddleware *middleware.AuthMiddleware,
) *router {
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.PrometheusMiddleware)
	return &router{
		router:          r,
		authHandler:     authHandler,
		healthHandler:   healthHandler,
		productHandler:  productHandler,
		categoryHandler: categoryHandler,
		shopHandler:     shopHandler,
		authMiddleware:  authMiddleware,
	}
}

func (r *router) RouteApp() *mux.Router {
	r.healthRoutes()
	r.authRoutes()
	r.productRoutes()
	r.categoryRoutes()
	r.metricsRoutes()
	r.shopRoutes()
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
	sub := r.router.PathPrefix("/products").Subrouter()
	// Apply auth middleware to all product mutation routes
	sub.Use(r.authMiddleware.Authenticate)
	sub.HandleFunc("", r.productHandler.Create).Methods(http.MethodPost)
	sub.HandleFunc("/{product_id}", r.productHandler.GetByID).Methods(http.MethodGet)
	sub.HandleFunc("/{product_id}", r.productHandler.Update).Methods(http.MethodPut)
	sub.HandleFunc("/{product_id}", r.productHandler.Delete).Methods(http.MethodDelete)
}

func (r *router) categoryRoutes() {
	sub := r.router.PathPrefix("/categories").Subrouter()
	// Apply auth middleware to all category routes
	sub.Use(r.authMiddleware.Authenticate)
	sub.HandleFunc("", r.categoryHandler.Create).Methods(http.MethodPost)
	sub.HandleFunc("/{category_id}", r.categoryHandler.GetByID).Methods(http.MethodGet)
	sub.HandleFunc("/{category_id}", r.categoryHandler.Update).Methods(http.MethodPut)
	sub.HandleFunc("/{category_id}", r.categoryHandler.Delete).Methods(http.MethodDelete)
}

func (r *router) shopRoutes() {
	// Public routes (no auth required)
	sub := r.router.PathPrefix("/shops").Subrouter()
	// GET /shops/{shop_id}/products?search=laptop&category_id=1&is_active=true&limit=20&cursor=0
	// Supports both filtered and non-filtered queries (backward compatible)
	// If no query params provided, returns all products with default pagination
	sub.HandleFunc("/{shop_id}/products", r.productHandler.GetAllByShopIDWithFilters).Methods(http.MethodGet)
	// GET /shops/{shop_id}/categories?search=electronics&sort=name&order=asc&limit=20&cursor=...
	sub.HandleFunc("/{shop_id}/categories", r.categoryHandler.GetAllByShopIDWithFilters).Methods(http.MethodGet)

	// Protected routes (auth required)
	protected := r.router.PathPrefix("/shops").Subrouter()
	protected.Use(r.authMiddleware.Authenticate)
	// GET /shops/{shop_id} - Get shop by ID (owner only)
	protected.HandleFunc("/{shop_id}", r.shopHandler.GetByID).Methods(http.MethodGet)
	// PUT /shops/{shop_id} - Update shop (owner only)
	protected.HandleFunc("/{shop_id}", r.shopHandler.Update).Methods(http.MethodPut)
}

func (r *router) metricsRoutes() {
	r.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
}
