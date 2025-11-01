package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router interface {
	RouteApp() *mux.Router
}
type router struct {
	router         *mux.Router
	authHandler    ports.AuthHandler
	healthHandler  ports.HealthHandler
	productHandler ports.ProductHandler
}

func NewRouter(authHandler ports.AuthHandler, healthHandler ports.HealthHandler, productHandler ports.ProductHandler) *router {
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.PrometheusMiddleware)
	return &router{
		router:         r,
		authHandler:    authHandler,
		healthHandler:  healthHandler,
		productHandler: productHandler,
	}
}

func (r *router) RouteApp() *mux.Router {
	r.healthRoutes()
	r.authRoutes()
	r.productRoutes()
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
}

func (r *router) shopRoutes() {
	sub := r.router.PathPrefix("/shops").Subrouter()
	sub.HandleFunc("/{shop_id}/products", r.productHandler.GetAllByShopID).Methods(http.MethodGet)
}

func (r *router) metricsRoutes() {
	r.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
}
