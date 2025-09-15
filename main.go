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
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/jwt"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/repositories/postgresql"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/server"
)

var Module = fx.Options(
	fx.Provide(
		// TOKEN
		fx.Annotate(jwt.NewTokenService, fx.As(new(ports.TokenService))),
		// AUTH
		fx.Annotate(http.NewAuthHandler, fx.As(new(ports.AuthHandler))),
		fx.Annotate(services.NewAuthService, fx.As(new(ports.AuthService))),

		// USER
		fx.Annotate(services.NewUserService, fx.As(new(ports.UserService))),
		fx.Annotate(postgresql.NewUserRepository, fx.As(new(ports.UserRepository))),

		// SHOP
		// fx.Annotate(services.NewShopService, fx.As(new(ports.ShopService))),
		fx.Annotate(postgresql.NewShopRepository, fx.As(new(ports.ShopRepository))),

		// ROLE
		fx.Annotate(postgresql.NewRoleRepository, fx.As(new(ports.RoleRepository))),

		// Sign UP
		fx.Annotate(services.NewSignupService, fx.As(new(ports.SignUpService))),
		fx.Annotate(postgresql.NewSignupRepository, fx.As(new(ports.SignupRepository))),

		fx.Annotate(auth.NewSignInUseCase, fx.As(new(ports.SignInUseCase))),
		fx.Annotate(auth.NewSignUpUseCase, fx.As(new(ports.SignUpUseCase))),

		// SERVER
		server.NewServer,
		fx.Annotate(server.NewRouter, fx.As(new(server.Router))),

		fx.Annotate(postgresql.NewDataBaseConnection, fx.As(new(postgresql.DataBaseConnection))),

		// fx.Annotate(handlers2.NewProductHandler, fx.As(new(handlers.ProductHandler))),
		// fx.Annotate(services.NewProductService, fx.As(new(iservices.ProductService))),
		// fx.Annotate(repositories.NewProductRepository, fx.As(new(persistence.ProductRepository))),

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
