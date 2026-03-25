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

func (j *TokenService) Generate(ctx context.Context, user *models.User, shopRoles []authclaims.ShopRole) (string, error) {
	if user == nil {
		return "", &errors.ValidationError{Message: errors.InvalidInput}
	}

	type tokenUser struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	userJSON, err := json.Marshal(tokenUser{ID: user.ID, Email: user.Email})
	if err != nil {
		return "", fmt.Errorf("failed to marshal user data: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user":  string(userJSON),
		"shops": shopRoles,
		"exp":   time.Now().Add(time.Minute * 30).Unix(),
		"iat":   time.Now().Unix(),
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
func (j *TokenService) parseTokenClaims(mapClaims map[string]interface{}) (*authclaims.TokenClaims, error) {
	user, err := j.parseUserFromClaims(mapClaims)
	if err != nil {
		return nil, err
	}

	return &authclaims.TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		ShopRoles: j.parseShopRolesFromClaims(mapClaims),
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

func (j *TokenService) parseShopRolesFromClaims(claims map[string]interface{}) []authclaims.ShopRole {
	rawShops, ok := claims["shops"].([]interface{})
	if !ok {
		return nil
	}

	var shopRoles []authclaims.ShopRole
	for _, raw := range rawShops {
		shopMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		floatID, idOk := shopMap["id"].(float64)
		role, roleOk := shopMap["role"].(string)
		if idOk && roleOk {
			sr := authclaims.ShopRole{
				ShopID: int(floatID),
				Role:   role,
			}
			// Parse permissions if present
			if rawPerms, ok := shopMap["permissions"].([]interface{}); ok {
				for _, rp := range rawPerms {
					if perm, ok := rp.(string); ok {
						sr.Permissions = append(sr.Permissions, perm)
					}
				}
			}
			shopRoles = append(shopRoles, sr)
		}
	}

	return shopRoles
}

func NewTokenService() *TokenService {
	return &TokenService{}
}
