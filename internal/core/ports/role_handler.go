package ports

import "net/http"

type RoleHandler interface {
	GetAssignable(http.ResponseWriter, *http.Request)
}
