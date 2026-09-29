package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines custom JWT claims
type Claims struct {
	UserID            string `json:"user_id"`
	Email             string `json:"email"`
	MustResetPassword bool   `json:"must_reset_password"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT string for a user
func GenerateToken(cfg *Config, userID, email string, mustReset bool) (string, error) {
	if cfg.JWTSecret == "" {
		return "", errors.New("jwt secret is not configured")
	}

	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpirationHours) * time.Hour)
	claims := &Claims{
		UserID:            userID,
		Email:             email,
		MustResetPassword: mustReset,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rutils-auth",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token string
func ValidateToken(cfg *Config, tokenStr string) (*Claims, error) {
	if cfg.JWTSecret == "" {
		return nil, errors.New("jwt secret is not configured")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}
