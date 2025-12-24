package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func NewJwtTokenGenerator(provider SecretProvider) TokenGenerator {
	secret := provider.Secret()
	return TokenGeneratorFunc(func(claims *VentClaims) (string, error) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString(secret)
	})
}

func NewJwtTokenAuthenticator(provider SecretProvider) TokenAuthenticator {
	secret := provider.Secret()
	return TokenAuthenticatorFunc(func(token string) (*VentClaims, error) {
		claims := &VentClaims{}
		t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			return secret, nil
		})
		if err != nil {
			return nil, err
		}
		if !t.Valid {
			return nil, fmt.Errorf("invalid token")
		}
		return claims, nil
	})
}
