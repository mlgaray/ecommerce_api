package jwt

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authclaims "github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
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
		"exp":      time.Now().Add(time.Minute * 30).Unix(),
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

	return j.parseTokenClaims(mapClaims)
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

// parseTokenClaims converts raw JWT MapClaims into a typed TokenClaims struct.
// This parsing logic lives in the adapter because it knows the JWT token format
// (e.g., "user" is a JSON-encoded string, "shop_ids" is []interface{} with float64 values).
func (j *TokenService) parseTokenClaims(mapClaims map[string]interface{}) (*authclaims.TokenClaims, error) {
	user, err := j.parseUserFromClaims(mapClaims)
	if err != nil {
		return nil, err
	}

	return &authclaims.TokenClaims{
		UserID:  user.ID,
		Email:   user.Email,
		ShopIDs: j.parseShopIDsFromClaims(mapClaims),
	}, nil
}

func (j *TokenService) parseUserFromClaims(claims map[string]interface{}) (*models.User, error) {
	userJSON, ok := claims["user"].(string)
	if !ok {
		return nil, &errors.AuthenticationError{Message: errors.TokenInvalid}
	}

	var user models.User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, &errors.AuthenticationError{Message: errors.TokenInvalid}
	}

	return &user, nil
}

func (j *TokenService) parseShopIDsFromClaims(claims map[string]interface{}) []int {
	var shopIDs []int
	if rawShopIDs, ok := claims["shop_ids"].([]interface{}); ok {
		for _, id := range rawShopIDs {
			if floatID, ok := id.(float64); ok {
				shopIDs = append(shopIDs, int(floatID))
			}
		}
	}
	return shopIDs
}

func NewTokenService() *TokenService {
	return &TokenService{}
}
