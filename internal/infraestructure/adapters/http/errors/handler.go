package errors

import (
	"encoding/json"
	"net/http"
	"reflect"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// errorStatusMap maps domain error types to HTTP status codes.
// This keeps HTTP knowledge in the infrastructure layer.
var errorStatusMap = map[reflect.Type]int{
	// HTTP Layer errors
	reflect.TypeOf(&BadRequestError{}):   http.StatusBadRequest,   // 400
	reflect.TypeOf(&UnauthorizedError{}): http.StatusUnauthorized, // 401
	reflect.TypeOf(&ForbiddenError{}):    http.StatusForbidden,    // 403

	// Domain errors
	reflect.TypeOf(&domainErrors.ValidationError{}):             http.StatusBadRequest,          // 400
	reflect.TypeOf(&domainErrors.AuthenticationError{}):         http.StatusUnauthorized,        // 401
	reflect.TypeOf(&domainErrors.AuthorizationError{}):          http.StatusForbidden,           // 403
	reflect.TypeOf(&domainErrors.RecordNotFoundError{}):         http.StatusNotFound,            // 404
	reflect.TypeOf(&domainErrors.DuplicateRecordError{}):        http.StatusConflict,            // 409
	reflect.TypeOf(&domainErrors.ReferentialIntegrityError{}):   http.StatusConflict,            // 409
	reflect.TypeOf(&domainErrors.ConcurrentModificationError{}): http.StatusConflict,            // 409
	reflect.TypeOf(&domainErrors.BusinessRuleError{}):           http.StatusUnprocessableEntity, // 422
}

// HandleError handles different error types and returns appropriate HTTP responses.
func HandleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	statusCode, message := mapErrorToHTTP(err)

	response := map[string]string{"error": message}
	responseData, encodeErr := json.Marshal(response)
	if encodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed_to_encode_response"}`))
		return
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(responseData)
}

// mapErrorToHTTP extracts status code and message from an error.
func mapErrorToHTTP(err error) (int, string) {
	errType := reflect.TypeOf(err)

	if statusCode, ok := errorStatusMap[errType]; ok {
		return statusCode, err.Error()
	}

	// Unknown error - log and return 500
	logs.WithFields(map[string]interface{}{
		"file":  "error_handler",
		"error": err.Error(),
	}).Error("Unhandled error")

	return http.StatusInternalServerError, "internal_server_error"
}
