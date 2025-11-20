package vent

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type VentClaims struct {
	jwt.RegisteredClaims
}

func NewClaims(userID string) VentClaims {
	now := time.Now()
	return VentClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
}

func CreateSignedToken(secret string, claims VentClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseSignedToken(secret string, signedToken string) (VentClaims, error) {
	claims := VentClaims{}
	token, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		return claims, err
	}
	if !token.Valid {
		return claims, fmt.Errorf("invalid token")
	}
	return claims, nil
}
