package auth

import (
	"sync/atomic"
	"testing"
)

func TestJwtUsesSecretProviderOnEachCall(t *testing.T) {
	var secret atomic.Value
	secret.Store([]byte("old-secret"))

	provider := SecretProviderFunc(func() []byte {
		return secret.Load().([]byte)
	})

	gen := NewJwtTokenGenerator(provider)
	authn := NewJwtTokenAuthenticator(provider)

	token, err := gen.Generate(NewClaims(42))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := authn.Authenticate(token)
	if err != nil {
		t.Fatalf("authenticate with original secret: %v", err)
	}
	if claims.Subject != "42" {
		t.Fatalf("subject = %q, want 42", claims.Subject)
	}

	// Rotate the secret after construction; generate/authenticate must pick it up.
	secret.Store([]byte("new-secret"))

	if _, err := authn.Authenticate(token); err == nil {
		t.Fatal("expected token signed with old secret to fail after rotation")
	}

	rotated, err := gen.Generate(NewClaims(7))
	if err != nil {
		t.Fatalf("generate after rotation: %v", err)
	}

	claims, err = authn.Authenticate(rotated)
	if err != nil {
		t.Fatalf("authenticate with rotated secret: %v", err)
	}
	if claims.Subject != "7" {
		t.Fatalf("subject = %q, want 7", claims.Subject)
	}
}
