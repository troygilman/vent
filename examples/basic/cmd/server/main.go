package main

import (
	"context"
	"log"
	"net/http"

	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"

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

	if err := seedAdminUser(ctx, client, credentialGenerator); err != nil {
		log.Fatalf("failed seeding admin user: %v", err)
	}
	if err := seedDemoData(ctx, client, credentialGenerator); err != nil {
		log.Fatalf("failed seeding demo data: %v", err)
	}

	adminHandler, err := admin.NewAdminHandler(admin.AdminConfig{
		Client: client,
		SecretProvider: auth.SecretProviderFunc(func() []byte {
			return []byte("secret")
		}),
		CredentialGenerator:     credentialGenerator,
		CredentialAuthenticator: auth.NewBCryptCredentialAuthenticator(),
		Schemas: admin.SchemaAdmins{
			User: UserAdmin{
				DefaultUserAdmin: admin.NewDefaultUserAdmin(client),
			},
			Author: AuthorAdmin{
				DefaultAuthorAdmin: admin.NewDefaultAuthorAdmin(client),
			},
			Book: BookAdmin{
				DefaultBookAdmin: admin.NewDefaultBookAdmin(client),
			},
			Review: ReviewAdmin{
				DefaultReviewAdmin: admin.NewDefaultReviewAdmin(client),
			},
		},
	})
	if err != nil {
		log.Fatalf("failed creating admin handler: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/admin/", adminHandler)
	log.Println("admin listening on http://localhost:8080/admin/")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
