package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
)

type Server struct {
	Router     Router
	httpServer *http.Server
}

func (s *Server) Initialize() {
	// CORS configuration - explicit settings to ensure headers are always added
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: false, // Cannot use credentials with AllowedOrigins: *
		MaxAge:           86400, // 24 hours - cache preflight requests
		Debug:            os.Getenv("ENVIRONMENT") == "development",
	})
	handler := corsMiddleware.Handler(s.Router.RouteApp())
	writeTimeout := 10 * time.Second // Producción
	if os.Getenv("ENVIRONMENT") == "test" {
		writeTimeout = 300 * time.Second // 5 minutos para debug
	}
	s.httpServer = &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func NewServer(router Router) *Server {
	return &Server{
		Router: router,
	}
}
