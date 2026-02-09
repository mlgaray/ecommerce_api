package ports

import "net/http"

// OrderHandler handles HTTP requests for order operations.
type OrderHandler interface {
	Create(http.ResponseWriter, *http.Request)
}
