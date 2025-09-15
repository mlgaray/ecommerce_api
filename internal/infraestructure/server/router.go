package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/middleware"
)

type Router interface {
	RouteApp() *mux.Router
}
type router struct {
	router      *mux.Router
	authHandler ports.AuthHandler
}

func NewRouter(authHandler ports.AuthHandler) *router {
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	// r.Use(middlewears.PrometheusMiddleware)
	return &router{
		router:      r,
		authHandler: authHandler,
	}
}

func (r *router) RouteApp() *mux.Router {
	r.authRoutes()
	return r.router
}

func (r *router) authRoutes() {
	sub := r.router.PathPrefix("/auth").Subrouter()
	sub.HandleFunc("/signin", r.authHandler.SignIn).Methods(http.MethodPost)
	sub.HandleFunc("/signup", r.authHandler.SignUp).Methods(http.MethodPost)
}
