package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/user"

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
	if err := seedDemoData(ctx, client); err != nil {
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
			Book: BookAdmin{
				DefaultBookAdmin: admin.NewDefaultBookAdmin(client),
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

func seedAdminUser(ctx context.Context, client *ent.Client, credentialGenerator auth.CredentialGenerator) error {
	exists, err := client.User.Query().Where(user.EmailEQ("admin@vent.com")).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	passwordHash, err := credentialGenerator.Generate("test_user")
	if err != nil {
		return err
	}

	_, err = client.User.Create().
		SetEmail("admin@vent.com").
		SetPasswordHash(passwordHash).
		SetIsStaff(true).
		SetIsSuperuser(true).
		Save(ctx)
	return err
}

func seedDemoData(ctx context.Context, client *ent.Client) error {
	if count, err := client.Book.Query().Count(ctx); err != nil {
		return err
	} else if count > 0 {
		return nil
	}

	fiction, err := client.Category.Create().
		SetName("Fiction").
		SetDescription("Novels and stories").
		Save(ctx)
	if err != nil {
		return err
	}

	author, err := client.Author.Create().
		SetName("Ada Lovelace").
		SetBio("Demo author").
		Save(ctx)
	if err != nil {
		return err
	}

	scifi, err := client.Tag.Create().SetName("sci-fi").Save(ctx)
	if err != nil {
		return err
	}
	classic, err := client.Tag.Create().SetName("classic").Save(ctx)
	if err != nil {
		return err
	}

	publishedAt := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	book, err := client.Book.Create().
		SetTitle("Analytical Engines").
		SetIsbn("978-0-000000-00-1").
		SetPages(320).
		SetPrice(24.99).
		SetPublished(true).
		SetPublishedAt(publishedAt).
		SetInternalNotes("Feature showcase seed book").
		SetAuthor(author).
		SetCategory(fiction).
		AddTags(scifi, classic).
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.Review.Create().
		SetReviewer("Casey").
		SetRating(5).
		SetBody("A clear tour of Vent's schema features.").
		SetBook(book).
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.AuditEvent.Create().
		SetAction("seed").
		SetDetail("Created demo library data").
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.ApiKey.Create().
		SetName("demo-key").
		SetToken("not-shown-in-admin").
		Save(ctx)
	return err
}
