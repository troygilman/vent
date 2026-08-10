package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func NewJwtTokenGenerator(provider SecretProvider) TokenGenerator {
	return TokenGeneratorFunc(func(claims *VentClaims) (string, error) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString(provider.Secret())
	})
}

func NewJwtTokenAuthenticator(provider SecretProvider) TokenAuthenticator {
	return TokenAuthenticatorFunc(func(token string) (*VentClaims, error) {
		claims := &VentClaims{}
		t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			return provider.Secret(), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			return nil, err
		}
		if !t.Valid {
			return nil, fmt.Errorf("invalid token")
		}
		return claims, nil
	})
}
