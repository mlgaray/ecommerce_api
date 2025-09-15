package jwt

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mlgaray/ecommerce_api/internal/core/entities"
	apperrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

var secretKey = "secret"

type TokenService struct{}

func (j *TokenService) Generate(ctx context.Context, user *models.User) (string, error) {
	if user == nil {
		return "", &apperrors.BadRequestError{Message: "user cannot be nil"}
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", &apperrors.InternalServiceError{Message: "failed to marshal user data"}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": string(userJSON),
		"exp":  time.Now().Add(time.Hour * 2).Unix(),
		"iat":  time.Now().Unix(),
	})

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", &apperrors.InternalServiceError{Message: "failed to sign token"}
	}

	return signedToken, nil
}

func (j *TokenService) VerifyToken(token string) (*entities.User, error) {
	if token == "" {
		return nil, &apperrors.BadRequestError{Message: "token cannot be empty"}
	}

	parse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, &apperrors.UnauthorizedError{Message: "unexpected signing method"}
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		// Comprueba si el error es del tipo jwt.TokenExpiredError
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &apperrors.UnauthorizedError{Message: "token is expired"}
		}
		return nil, &apperrors.UnauthorizedError{Message: "could not parse token"}
	}

	if !parse.Valid {
		return nil, &apperrors.UnauthorizedError{Message: "token is not valid"}
	}

	// claims, ok := parse.Claims.(jwt.MapClaims)
	/*
		if !ok {
			return nil, &apperrors.UnauthorizedError{Message: "could not get claims"}
		}*/

	// email := claims["email"].(string)
	return nil, nil
}

func NewTokenService() *TokenService {
	return &TokenService{}
}
