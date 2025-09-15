package errors

import (
	"encoding/json"
	"net/http"
)

// HandleError handles different error types and returns appropriate HTTP responses
func HandleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var statusCode int
	var message string

	switch e := err.(type) {
	case *NotFoundError:
		statusCode = http.StatusNotFound
		message = e.Message
	case *UnauthorizedError:
		statusCode = http.StatusUnauthorized
		message = e.Message
	case *ForbiddenError:
		statusCode = http.StatusForbidden
		message = e.Message
	case *BadRequestError:
		statusCode = http.StatusBadRequest
		message = e.Message
	case *ConflictError:
		statusCode = http.StatusConflict
		message = e.Message
	case *InternalServiceError:
		statusCode = http.StatusInternalServerError
		message = e.Message
	default:
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	}

	response := map[string]string{"error": message}

	// Codificar la respuesta antes de escribir headers
	responseData, encodeErr := json.Marshal(response)
	if encodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"Failed to encode response"}`))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(responseData)
}
