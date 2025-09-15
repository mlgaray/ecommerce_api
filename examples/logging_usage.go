package examples

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Ejemplo 1: Uso directo del logger global
func ExampleGlobalLogger() {
	logs.Info("Application started")
	logs.Error("Something went wrong")

	logs.WithFields(map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}).Info("User logged in")
}

// Ejemplo 2: Uso con context (en HTTP handlers/repositories)
func ExampleContextLogger(ctx context.Context) {
	// El context ya tiene el logger con request_id, path, etc.
	logs.FromContext(ctx).WithFields(map[string]interface{}{
		"database": "postgresql",
		"query":    "SELECT * FROM users",
	}).Info("Executing database query")

	// Para errores con contexto
	logs.FromContext(ctx).WithFields(map[string]interface{}{
		"error":   "connection timeout",
		"timeout": "30s",
	}).Error("Database connection failed")
}

// Ejemplo 3: Uso directo en repositories (recomendado)
func ExampleRepositoryDirect(ctx context.Context, email string) error {
	// Logger directo desde infraestructura
	logger := logs.FromContext(ctx).WithFields(map[string]interface{}{
		"operation": "create_user",
		"email":     email,
	})

	logger.Info("Starting user creation")

	// Simulamos un error
	err := doSomething()
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to create user in database")
		return err
	}

	logger.Info("User created successfully")
	return nil
}

func doSomething() error {
	return nil
}
