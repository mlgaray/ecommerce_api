package ports

import (
	"net/http"
)

type HealthHandler interface {
	Health(w http.ResponseWriter, r *http.Request)
}
