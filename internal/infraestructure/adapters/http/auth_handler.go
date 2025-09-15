package http

import (
	"encoding/json"
	"net/http"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
)

type AuthHandler struct {
	signIn ports.SignInUseCase
	signUp ports.SignUpUseCase
}

func (u *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req contracts.SignInRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errors.HandleError(w, &errors.BadRequestError{Message: "invalid_JSON_format"})
		return
	}

	if err := req.Validate(); err != nil {
		errors.HandleError(w, err)
		return
	}

	user := req.ToUser()
	token, err := u.signIn.Execute(ctx, user)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response := contracts.SignInResponse{Token: token}

	responseData, err := json.Marshal(response)
	if err != nil {
		errors.HandleError(w, &errors.InternalServiceError{Message: "Failed to encode response"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

func (u *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req contracts.SignUpRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		errors.HandleError(w, &errors.BadRequestError{Message: "Invalid JSON format"})
		return
	}

	if err := req.Validate(); err != nil {
		errors.HandleError(w, err)
		return
	}

	err = u.signUp.Execute(ctx, &req.User, &req.Shop)
	if err != nil {
		errors.HandleError(w, err)
		return
	}
	response := map[string]int{"status": http.StatusOK}

	// Encode response before writing headers
	responseData, err := json.Marshal(response)
	if err != nil {
		errors.HandleError(w, &errors.InternalServiceError{Message: "Failed to encode response"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

func NewAuthHandler(signIn ports.SignInUseCase, signUp ports.SignUpUseCase) *AuthHandler {
	return &AuthHandler{
		signUp: signUp,
		signIn: signIn,
	}
}
