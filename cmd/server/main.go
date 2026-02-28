package main

import (
	"context"
	"log"
	"net/http"

	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/ent"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:tmp/test.db?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	credentialGenerator := auth.NewBCryptCredentialGenerator()

	passwordHash, err := credentialGenerator.Generate("test_user")
	if err != nil {
		panic(err)
	}

	client.AuthUser.Create().
		SetEmail("admin@vent.com").
		SetPasswordHash(passwordHash).
		SetIsStaff(true).
		SetIsSuperuser(true).
		Save(ctx)

	mux := http.NewServeMux()
	mux.Handle("/admin/", ent.NewAdminHandler(ent.AdminConfig{
		Client: client,
		SecretProvider: auth.SecretProviderFunc(func() []byte {
			return []byte("secret")
		}),
		CredentialGenerator:     credentialGenerator,
		CredentialAuthenticator: auth.NewBCryptCredentialAuthenticator(),
	}))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
