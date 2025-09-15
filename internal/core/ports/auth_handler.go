package ports

import (
	"net/http"
)

type AuthHandler interface {
	SignIn(w http.ResponseWriter, r *http.Request)
	SignUp(w http.ResponseWriter, r *http.Request)
}
