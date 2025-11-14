package jwt

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mlgaray/ecommerce_api/internal/core/entities"
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

var secretKey = "secret"

type TokenService struct{}

func (j *TokenService) Generate(ctx context.Context, user *models.User) (string, error) {
	if user == nil {
		return "", &errors.ValidationError{Message: errors.InvalidInput}
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user data: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": string(userJSON),
		"exp":  time.Now().Add(time.Hour * 2).Unix(),
		"iat":  time.Now().Unix(),
	})

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (j *TokenService) VerifyToken(token string) (*entities.User, error) {
	if token == "" {
		return nil, &errors.ValidationError{Message: errors.TokenCannotBeEmpty}
	}

	parse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, &errors.AuthenticationError{Message: errors.UnexpectedSigningMethod}
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		// Comprueba si el error es del tipo jwt.TokenExpiredError
		if stderrors.Is(err, jwt.ErrTokenExpired) {
			return nil, &errors.AuthenticationError{Message: errors.TokenExpired}
		}
		return nil, &errors.AuthenticationError{Message: errors.CouldNotParseToken}
	}

	if !parse.Valid {
		return nil, &errors.AuthenticationError{Message: errors.TokenInvalid}
	}

	// claims, ok := parse.Claims.(jwt.MapClaims)
	/*
		if !ok {
			return nil, &errors.AuthenticationError{Message: "could not get claims"}
		}*/

	// email := claims["email"].(string)
	return nil, nil
}

func NewTokenService() *TokenService {
	return &TokenService{}
}
