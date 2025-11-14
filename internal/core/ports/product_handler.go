package ports

import "net/http"

type ProductHandler interface {
	Create(http.ResponseWriter, *http.Request)
	GetAllByShopID(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
}
