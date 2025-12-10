package errors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		// HTTP Layer errors
		{
			name:           "BadRequestError returns 400",
			err:            &BadRequestError{Message: "invalid_input"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid_input"}`,
		},
		{
			name:           "UnauthorizedError returns 401",
			err:            &UnauthorizedError{Message: "invalid_token"},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid_token"}`,
		},
		{
			name:           "ForbiddenError returns 403",
			err:            &ForbiddenError{Message: "access_denied"},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"access_denied"}`,
		},

		// Domain errors
		{
			name:           "ValidationError returns 400",
			err:            &domainErrors.ValidationError{Message: "field_required"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"field_required"}`,
		},
		{
			name:           "AuthenticationError returns 401",
			err:            &domainErrors.AuthenticationError{Message: "token_expired"},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"token_expired"}`,
		},
		{
			name:           "AuthorizationError returns 403",
			err:            &domainErrors.AuthorizationError{Message: "access_denied"},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"access_denied"}`,
		},
		{
			name:           "RecordNotFoundError returns 404",
			err:            &domainErrors.RecordNotFoundError{Message: "category_not_found"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"category_not_found"}`,
		},
		{
			name:           "DuplicateRecordError returns 409",
			err:            &domainErrors.DuplicateRecordError{Message: "already_exists"},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"already_exists"}`,
		},
		{
			name:           "ReferentialIntegrityError returns 409",
			err:            &domainErrors.ReferentialIntegrityError{Message: "category_has_products"},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"category_has_products"}`,
		},
		{
			name:           "BusinessRuleError returns 422",
			err:            &domainErrors.BusinessRuleError{Message: "invalid_operation"},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"invalid_operation"}`,
		},

		// Unknown errors
		{
			name:           "Unknown error returns 500",
			err:            errors.New("unexpected error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal_server_error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			w := httptest.NewRecorder()

			// Act
			HandleError(w, tt.err)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestMapErrorToHTTP(t *testing.T) {
	t.Run("returns correct status and message for known error", func(t *testing.T) {
		err := &domainErrors.RecordNotFoundError{Message: "not_found"}

		statusCode, message := mapErrorToHTTP(err)

		assert.Equal(t, http.StatusNotFound, statusCode)
		assert.Equal(t, "not_found", message)
	})

	t.Run("returns 500 and generic message for unknown error", func(t *testing.T) {
		err := errors.New("some technical error")

		statusCode, message := mapErrorToHTTP(err)

		assert.Equal(t, http.StatusInternalServerError, statusCode)
		assert.Equal(t, "internal_server_error", message)
	})
}
