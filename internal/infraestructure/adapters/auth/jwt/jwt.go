package jwt

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	authclaims "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/claims"
)

var secretKey = "secret"

type TokenService struct{}

func (j *TokenService) Generate(ctx context.Context, user *models.User, shopIDs []int) (string, error) {
	if user == nil {
		return "", &errors.ValidationError{Message: errors.InvalidInput}
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user data: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user":     string(userJSON),
		"shop_ids": shopIDs,
		"exp":      time.Now().Add(time.Minute * 10).Unix(),
		"iat":      time.Now().Unix(),
	})

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateAndParseClaims validates a JWT token and extracts the claims.
func (j *TokenService) ValidateAndParseClaims(tokenString string) (*authclaims.TokenClaims, error) {
	if tokenString == "" {
		return nil, &errors.AuthenticationError{Message: errors.TokenCannotBeEmpty}
	}

	token, err := j.parseAndValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	mapClaims, err := j.extractMapClaims(token)
	if err != nil {
		return nil, err
	}

	return authclaims.NewTokenClaims(mapClaims)
}

func (j *TokenService) parseAndValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &errors.AuthenticationError{Message: errors.UnexpectedSigningMethod}
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if stderrors.Is(err, jwt.ErrTokenExpired) {
			return nil, &errors.AuthenticationError{Message: errors.TokenExpired}
		}
		return nil, &errors.AuthenticationError{Message: errors.CouldNotParseToken}
	}

	if !token.Valid {
		return nil, &errors.AuthenticationError{Message: errors.TokenInvalid}
	}

	return token, nil
}

func (j *TokenService) extractMapClaims(token *jwt.Token) (map[string]interface{}, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &errors.AuthenticationError{Message: errors.CouldNotParseToken}
	}
	return claims, nil
}

func NewTokenService() *TokenService {
	return &TokenService{}
}
