package ports

import "net/http"

type ProductHandler interface {
	Create(http.ResponseWriter, *http.Request)
	GetAllByShopIDWithFilters(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}
