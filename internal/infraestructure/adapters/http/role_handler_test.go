package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// NewRoleHandler Tests
// =============================================================================

func TestNewRoleHandler(t *testing.T) {
	t.Run("creates handler with dependency", func(t *testing.T) {
		// Arrange
		ucMock := mocks.NewGetAssignableRolesUseCase(t)

		// Act
		handler := NewRoleHandler(ucMock)

		// Assert
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.getAssignableRoles)
	})

	t.Run("creates handler with nil dependency", func(t *testing.T) {
		// Act
		handler := NewRoleHandler(nil)

		// Assert
		assert.NotNil(t, handler)
	})
}

// =============================================================================
// GetAssignable Tests
// =============================================================================

func TestRoleHandler_GetAssignable(t *testing.T) {
	t.Run("when use case returns roles then returns 200 with roles", func(t *testing.T) {
		// Arrange
		roles := []*models.Role{
			{ID: 2, Name: "admin", Description: "Administrador de la tienda con permisos operations completos"},
			{ID: 3, Name: "encargado", Description: "Encargado de tienda con permisos de products, categorías y órdenes"},
		}

		ucMock := mocks.NewGetAssignableRolesUseCase(t)
		ucMock.EXPECT().Execute(mock.Anything).Return(roles, nil)

		handler := NewRoleHandler(ucMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/roles", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.GetAssignable(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body responses.ListRolesResponse
		err := json.Unmarshal(rec.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Len(t, body.Roles, 2)
		assert.Equal(t, "admin", body.Roles[0].Name)
		assert.Equal(t, 3, body.Roles[1].ID)
	})

	t.Run("when use case returns no roles then returns 200 with empty list", func(t *testing.T) {
		// Arrange
		ucMock := mocks.NewGetAssignableRolesUseCase(t)
		ucMock.EXPECT().Execute(mock.Anything).Return([]*models.Role{}, nil)

		handler := NewRoleHandler(ucMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/roles", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.GetAssignable(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		var body responses.ListRolesResponse
		err := json.Unmarshal(rec.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Empty(t, body.Roles)
	})

	t.Run("when use case returns error then returns 500", func(t *testing.T) {
		// Arrange
		ucMock := mocks.NewGetAssignableRolesUseCase(t)
		ucMock.EXPECT().Execute(mock.Anything).Return(nil, errors.New("unexpected error"))

		handler := NewRoleHandler(ucMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/roles", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.GetAssignable(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
