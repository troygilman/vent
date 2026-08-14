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

func seedUser(ctx context.Context, client *ent.Client, credentialGenerator auth.CredentialGenerator, email, password string, staff bool) (*ent.User, error) {
	existing, err := client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	passwordHash, err := credentialGenerator.Generate(password)
	if err != nil {
		return nil, err
	}
	return client.User.Create().
		SetEmail(email).
		SetPasswordHash(passwordHash).
		SetIsStaff(staff).
		Save(ctx)
}

func seedDemoData(ctx context.Context, client *ent.Client, credentialGenerator auth.CredentialGenerator) error {
	if count, err := client.Book.Query().Count(ctx); err != nil {
		return err
	} else if count > 0 {
		return nil
	}

	ada, err := seedUser(ctx, client, credentialGenerator, "ada@vent.com", "test_user", true)
	if err != nil {
		return err
	}
	charles, err := seedUser(ctx, client, credentialGenerator, "charles@vent.com", "test_user", true)
	if err != nil {
		return err
	}
	casey, err := seedUser(ctx, client, credentialGenerator, "casey@vent.com", "test_user", false)
	if err != nil {
		return err
	}
	riley, err := seedUser(ctx, client, credentialGenerator, "riley@vent.com", "test_user", false)
	if err != nil {
		return err
	}

	author, err := client.Author.Create().
		SetUser(ada).
		SetActive(true).
		Save(ctx)
	if err != nil {
		return err
	}
	inactiveAuthor, err := client.Author.Create().
		SetUser(charles).
		SetActive(false).
		Save(ctx)
	if err != nil {
		return err
	}

	publishedAt := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	book, err := client.Book.Create().
		SetTitle("Analytical Engines").
		SetPages(320).
		SetPublished(true).
		SetPublishedAt(publishedAt).
		SetInternalNotes("Feature showcase seed book").
		SetAuthor(author).
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.Book.Create().
		SetTitle("Notes on the Difference Engine").
		SetPages(48).
		SetPublished(false).
		SetAuthor(inactiveAuthor).
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.Review.Create().
		SetUser(casey).
		SetRating(5).
		SetBody("A clear tour of Vent's schema features.").
		SetBook(book).
		Save(ctx)
	if err != nil {
		return err
	}
	_, err = client.Review.Create().
		SetUser(riley).
		SetRating(3).
		SetBody("Useful, but still a draft.").
		SetBook(book).
		Save(ctx)
	return err
}
