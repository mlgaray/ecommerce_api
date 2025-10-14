package ports

import "net/http"

type ProductHandler interface {
	Create(http.ResponseWriter, *http.Request)
}
