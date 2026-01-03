package ports

import "net/http"

type ShopHandler interface {
	GetByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
}
