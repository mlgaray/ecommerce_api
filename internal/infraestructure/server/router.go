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
	router        *mux.Router
	authHandler   ports.AuthHandler
	healthHandler ports.HealthHandler
}

func NewRouter(authHandler ports.AuthHandler, healthHandler ports.HealthHandler) *router {
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.PrometheusMiddleware)
	return &router{
		router:        r,
		authHandler:   authHandler,
		healthHandler: healthHandler,
	}
}

func (r *router) RouteApp() *mux.Router {
	r.healthRoutes()
	r.authRoutes()
	r.metricsRoutes()
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

func (r *router) metricsRoutes() {
	r.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
}
