package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// SignIn Tests
// =============================================================================

func TestAuthHandler_SignIn(t *testing.T) {
	t.Run("when valid credentials then returns 200 with token", func(t *testing.T) {
		// Arrange
		signInMock := mocks.NewSignInUseCase(t)
		signInMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.User")).
			Return("valid.jwt.token", nil)

		handler := NewAuthHandler(signInMock, nil)

		body := contracts.SignInRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignIn(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response contracts.SignInResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "valid.jwt.token", response.Token)
	})

	t.Run("when invalid JSON then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignIn(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when email is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		body := contracts.SignInRequest{
			Email:    "",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignIn(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when password is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		body := contracts.SignInRequest{
			Email:    "test@example.com",
			Password: "",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignIn(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when invalid credentials then returns 401", func(t *testing.T) {
		// Arrange
		signInMock := mocks.NewSignInUseCase(t)
		signInMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.User")).
			Return("", &domainErrors.AuthenticationError{Message: domainErrors.InvalidUserCredentials})

		handler := NewAuthHandler(signInMock, nil)

		body := contracts.SignInRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignIn(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("when user not found then returns 404", func(t *testing.T) {
		// Arrange
		signInMock := mocks.NewSignInUseCase(t)
		signInMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.User")).
			Return("", &domainErrors.RecordNotFoundError{Message: domainErrors.UserNotFound})

		handler := NewAuthHandler(signInMock, nil)

		body := contracts.SignInRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignIn(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when empty body then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignIn(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// SignUp Tests
// =============================================================================

func TestAuthHandler_SignUp(t *testing.T) {
	t.Run("when valid data then returns 200", func(t *testing.T) {
		// Arrange
		signUpMock := mocks.NewSignUpUseCase(t)
		signUpMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.User"), mock.AnythingOfType("*models.Shop")).
			Return(nil)

		handler := NewAuthHandler(nil, signUpMock)

		body := contracts.SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john@example.com",
				Password: "password123",
				Phone:    "+54 11 1234-5678",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+54 11 9876-5432",
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]int
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response["status"])
	})

	t.Run("when invalid JSON then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when email is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		body := contracts.SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "",
				Password: "password123",
			},
			Shop: models.Shop{
				Name: "John's Shop",
				Slug: "johns-shop",
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when password is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		body := contracts.SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john@example.com",
				Password: "",
			},
			Shop: models.Shop{
				Name: "John's Shop",
				Slug: "johns-shop",
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when user already exists then returns 409", func(t *testing.T) {
		// Arrange
		signUpMock := mocks.NewSignUpUseCase(t)
		signUpMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.User"), mock.AnythingOfType("*models.Shop")).
			Return(&domainErrors.DuplicateRecordError{Message: domainErrors.UserAlreadyExists})

		handler := NewAuthHandler(nil, signUpMock)

		body := contracts.SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "existing@example.com",
				Password: "password123",
				Phone:    "+54 11 1234-5678",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+54 11 9876-5432",
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("when shop name is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		body := contracts.SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			Shop: models.Shop{
				Name: "",
				Slug: "johns-shop",
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop slug is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewAuthHandler(nil, nil)

		body := contracts.SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			Shop: models.Shop{
				Name: "John's Shop",
				Slug: "",
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when validation error then returns 400", func(t *testing.T) {
		// Arrange
		signUpMock := mocks.NewSignUpUseCase(t)
		signUpMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.User"), mock.AnythingOfType("*models.Shop")).
			Return(&domainErrors.ValidationError{Message: "invalid_input"})

		handler := NewAuthHandler(nil, signUpMock)

		body := contracts.SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john@example.com",
				Password: "password123",
				Phone:    "+54 11 1234-5678",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+54 11 9876-5432",
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		handler.SignUp(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// NewAuthHandler Tests
// =============================================================================

func TestNewAuthHandler(t *testing.T) {
	t.Run("creates handler with dependencies", func(t *testing.T) {
		// Arrange
		signInMock := mocks.NewSignInUseCase(t)
		signUpMock := mocks.NewSignUpUseCase(t)

		// Act
		handler := NewAuthHandler(signInMock, signUpMock)

		// Assert
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.signIn)
		assert.NotNil(t, handler.signUp)
	})

	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		// Act
		handler := NewAuthHandler(nil, nil)

		// Assert
		assert.NotNil(t, handler)
		assert.Nil(t, handler.signIn)
		assert.Nil(t, handler.signUp)
	})
}
