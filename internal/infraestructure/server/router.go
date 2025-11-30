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
}

func NewRouter(authHandler ports.AuthHandler, healthHandler ports.HealthHandler, productHandler ports.ProductHandler, categoryHandler ports.CategoryHandler) *router {
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.PrometheusMiddleware)
	return &router{
		router:          r,
		authHandler:     authHandler,
		healthHandler:   healthHandler,
		productHandler:  productHandler,
		categoryHandler: categoryHandler,
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
	sub.HandleFunc("", r.productHandler.Create).Methods(http.MethodPost)
	sub.HandleFunc("/{product_id}", r.productHandler.GetByID).Methods(http.MethodGet)
	sub.HandleFunc("/{product_id}", r.productHandler.Update).Methods(http.MethodPut)
}

func (r *router) categoryRoutes() {
	sub := r.router.PathPrefix("/categories").Subrouter()
	sub.HandleFunc("", r.categoryHandler.Create).Methods(http.MethodPost)
}

func (r *router) shopRoutes() {
	sub := r.router.PathPrefix("/shops").Subrouter()
	// GET /shops/{shop_id}/products?search=laptop&category_id=1&is_active=true&limit=20&cursor=0
	// Supports both filtered and non-filtered queries (backward compatible)
	// If no query params provided, returns all products with default pagination
	sub.HandleFunc("/{shop_id}/products", r.productHandler.GetAllByShopIDWithFilters).Methods(http.MethodGet)
}

func (r *router) metricsRoutes() {
	r.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
}
