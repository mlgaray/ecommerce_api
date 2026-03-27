package jwt

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	authclaims "github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

func TestTokenService_Generate(t *testing.T) {
	t.Run("when user is valid then generates token", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		user := &models.User{ID: 1, Email: "test@example.com"}
		shopRoles := []authclaims.ShopRole{{ShopID: 1, Role: "owner"}, {ShopID: 2, Role: "admin"}}

		// Act
		token, err := service.Generate(context.Background(), user, shopRoles)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("when user is nil then returns ValidationError", func(t *testing.T) {
		// Arrange
		service := NewTokenService()

		// Act
		token, err := service.Generate(context.Background(), nil, []authclaims.ShopRole{{ShopID: 1, Role: "owner"}})

		// Assert
		assert.Empty(t, token)
		assert.Error(t, err)
		validationErr, ok := err.(*errors.ValidationError)
		assert.True(t, ok)
		assert.Equal(t, errors.InvalidInput, validationErr.Message)
	})

	t.Run("when shopRoles is empty then generates token", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		user := &models.User{ID: 1, Email: "test@example.com"}

		// Act
		token, err := service.Generate(context.Background(), user, []authclaims.ShopRole{})

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("when shopRoles is nil then generates token", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		user := &models.User{ID: 1, Email: "test@example.com"}

		// Act
		token, err := service.Generate(context.Background(), user, nil)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("when user has password then token payload does not contain password", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		user := &models.User{ID: 1, Email: "test@example.com", Password: "super-secret-password"}
		shopRoles := []authclaims.ShopRole{{ShopID: 1, Role: "owner"}}

		// Act
		token, err := service.Generate(context.Background(), user, shopRoles)

		// Assert
		assert.NoError(t, err)

		// Decode the JWT payload (second segment) and verify no password field
		parts := strings.Split(token, ".")
		assert.Len(t, parts, 3)

		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		assert.NoError(t, err)

		payloadStr := string(payload)
		assert.NotContains(t, payloadStr, "password")
		assert.NotContains(t, payloadStr, "super-secret-password")
	})
}

func TestTokenService_ValidateAndParseClaims(t *testing.T) {
	t.Run("when token is valid then returns claims", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		user := &models.User{ID: 1, Email: "test@example.com"}
		shopRoles := []authclaims.ShopRole{{ShopID: 1, Role: "owner"}, {ShopID: 2, Role: "admin"}}
		token, _ := service.Generate(context.Background(), user, shopRoles)

		// Act
		parsedClaims, err := service.ValidateAndParseClaims(token)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, parsedClaims)
		assert.Equal(t, 1, parsedClaims.UserID)
		assert.Equal(t, "test@example.com", parsedClaims.Email)
		assert.Equal(t, shopRoles, parsedClaims.ShopRoles)
	})

	t.Run("when token is empty then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		service := NewTokenService()

		// Act
		claims, err := service.ValidateAndParseClaims("")

		// Assert
		assert.Nil(t, claims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.TokenCannotBeEmpty, authErr.Message)
	})

	t.Run("when token is malformed then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		service := NewTokenService()

		// Act
		claims, err := service.ValidateAndParseClaims("invalid.token.here")

		// Assert
		assert.Nil(t, claims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.CouldNotParseToken, authErr.Message)
	})

	t.Run("when token is expired then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		// Create an expired token manually
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user":  `{"id":1,"email":"test@example.com"}`,
			"shops": []interface{}{map[string]interface{}{"id": float64(1), "role": "owner"}},
			"exp":   time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
			"iat":   time.Now().Add(-2 * time.Hour).Unix(),
		})
		tokenString, _ := expiredToken.SignedString([]byte(secretKey))

		// Act
		claims, err := service.ValidateAndParseClaims(tokenString)

		// Assert
		assert.Nil(t, claims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.TokenExpired, authErr.Message)
	})

	t.Run("when token has wrong signing method then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		service := NewTokenService()

		// Create a valid-looking JWT with RS256 algorithm in header
		// The keyfunc will reject it because we only accept HMAC
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(
			fmt.Sprintf(`{"user":"{\"id\":1,\"email\":\"test@example.com\"}","shops":[{"id":1,"role":"owner"}],"exp":%d}`,
				time.Now().Add(time.Hour).Unix())))
		// Add a valid base64 signature (content doesn't matter, just needs to be parseable)
		signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature-bytes"))

		tokenString := header + "." + payload + "." + signature

		// Act
		claims, err := service.ValidateAndParseClaims(tokenString)

		// Assert
		assert.Nil(t, claims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		// The error could be either unexpected_signing_method or could_not_parse_token
		// depending on when the JWT library fails
		assert.Contains(t, []string{errors.UnexpectedSigningMethod, errors.CouldNotParseToken}, authErr.Message)
	})

	t.Run("when token is signed with wrong key then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user":  `{"id":1,"email":"test@example.com"}`,
			"shops": []interface{}{map[string]interface{}{"id": float64(1), "role": "owner"}},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("wrong-secret-key"))

		// Act
		claims, err := service.ValidateAndParseClaims(tokenString)

		// Assert
		assert.Nil(t, claims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.CouldNotParseToken, authErr.Message)
	})

	t.Run("when token has invalid user claim then returns AuthenticationError", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user":  "invalid-json",
			"shops": []interface{}{map[string]interface{}{"id": float64(1), "role": "owner"}},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(secretKey))

		// Act
		claims, err := service.ValidateAndParseClaims(tokenString)

		// Assert
		assert.Nil(t, claims)
		assert.Error(t, err)
		authErr, ok := err.(*errors.AuthenticationError)
		assert.True(t, ok)
		assert.Equal(t, errors.TokenInvalid, authErr.Message)
	})

	t.Run("when token has no shop_ids then returns claims with empty slice", func(t *testing.T) {
		// Arrange
		service := NewTokenService()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user": `{"id":1,"email":"test@example.com"}`,
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(secretKey))

		// Act
		claims, err := service.ValidateAndParseClaims(tokenString)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, 1, claims.UserID)
		assert.Empty(t, claims.ShopRoles)
	})
}

func TestNewTokenService(t *testing.T) {
	t.Run("creates new token service instance", func(t *testing.T) {
		// Act
		service := NewTokenService()

		// Assert
		assert.NotNil(t, service)
	})
}
