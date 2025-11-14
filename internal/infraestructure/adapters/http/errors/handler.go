package errors

import (
	"encoding/json"
	"net/http"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// HandleError handles different error types and returns appropriate HTTP responses
// This function maps domain errors to HTTP status codes
func HandleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var statusCode int
	var message string

	// Map domain and HTTP errors to HTTP status codes
	switch e := err.(type) {
	// HTTP Layer errors (400 Bad Request)
	case *BadRequestError:
		statusCode = http.StatusBadRequest
		message = e.Message

	// Domain errors mapped to HTTP status codes
	case *domainErrors.RecordNotFoundError:
		statusCode = http.StatusNotFound // 404
		message = e.Message

	case *domainErrors.DuplicateRecordError:
		statusCode = http.StatusConflict // 409
		message = e.Message

	case *domainErrors.ValidationError:
		statusCode = http.StatusBadRequest // 400
		message = e.Message

	case *domainErrors.AuthenticationError:
		statusCode = http.StatusUnauthorized // 401
		message = e.Message

	case *domainErrors.AuthorizationError:
		statusCode = http.StatusForbidden // 403
		message = e.Message

	case *domainErrors.BusinessRuleError:
		statusCode = http.StatusUnprocessableEntity // 422
		message = e.Message

	default:
		// Any other error (technical, unexpected) = 500
		// Do not expose technical details to the client
		statusCode = http.StatusInternalServerError
		message = "internal_server_error"
		logs.WithFields(map[string]interface{}{
			"file":  "error_handler",
			"error": err.Error(),
		}).Error("Unhandled error")
	}

	response := map[string]string{"error": message}

	// Encode response before writing headers
	responseData, encodeErr := json.Marshal(response)
	if encodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed_to_encode_response"}`))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(responseData)
}
