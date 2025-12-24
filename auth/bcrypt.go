package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func NewBCryptCredentialGenerator() CredentialGenerator {
	return CredentialGeneratorFunc(func(password string) (string, error) {
		bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		return string(bytes), err
	})
}

func NewBCryptCredentialAuthenticator() CredentialAuthenticator {
	return CredentialAuthenticatorFunc(func(password, hash string) error {
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
			return err
		}
		return nil
	})
}
