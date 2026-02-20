package ports

import "net/http"

// OrderHandler handles HTTP requests for order operations.
type OrderHandler interface {
	Create(http.ResponseWriter, *http.Request)
	GetAll(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	UpdateStatus(http.ResponseWriter, *http.Request)
}
