package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type VentClaims struct {
	jwt.RegisteredClaims
}

func NewClaims(userID int) *VentClaims {
	now := time.Now()
	return &VentClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
}

type SecretProvider interface {
	Secret() []byte
}

type SecretProviderFunc func() []byte

func (f SecretProviderFunc) Secret() []byte {
	return f()
}

type TokenGenerator interface {
	Generate(*VentClaims) (string, error)
}

type TokenGeneratorFunc func(claims *VentClaims) (string, error)

func (f TokenGeneratorFunc) Generate(claims *VentClaims) (string, error) {
	return f(claims)
}

type TokenAuthenticator interface {
	Authenticate(token string) (*VentClaims, error)
}

type TokenAuthenticatorFunc func(token string) (*VentClaims, error)

func (f TokenAuthenticatorFunc) Authenticate(token string) (*VentClaims, error) {
	return f(token)
}
