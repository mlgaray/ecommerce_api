package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Auth handler log field constants
const (
	AuthHandlerField = "auth_handler"
	SignInFunction   = "sign_in"
	SignUpFunction   = "sign_up"
)

type AuthHandler struct {
	signIn ports.SignInUseCase
	signUp ports.SignUpUseCase
}

func (u *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req contracts.SignInRequest

	// Parse JSON request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_json_format"})
		return
	}

	// Validate HTTP input
	if err := req.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute business logic
	user := req.ToUser()
	token, err := u.signIn.Execute(ctx, user)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Build response
	response := contracts.SignInResponse{Token: token}
	responseData, err := json.Marshal(response)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     AuthHandlerField,
			"function": SignInFunction,
			"sub_func": "json.Marshal",
			"error":    err.Error(),
		}).Error("Failed to encode response")
		httpErrors.HandleError(w, fmt.Errorf("failed to encode response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

func (u *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req contracts.SignUpRequest

	// Parse JSON request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_json_format"})
		return
	}

	// Validate HTTP input
	if err := req.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute business logic
	err = u.signUp.Execute(ctx, &req.User, &req.Shop)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Build response
	response := map[string]int{"status": http.StatusOK}
	responseData, err := json.Marshal(response)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     AuthHandlerField,
			"function": SignUpFunction,
			"sub_func": "json.Marshal",
			"error":    err.Error(),
		}).Error("Failed to encode response")
		httpErrors.HandleError(w, fmt.Errorf("failed to encode response"))
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
