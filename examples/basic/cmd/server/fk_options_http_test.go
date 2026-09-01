package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/troygilman/vent"
	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

func TestReviewAddFormCapsBookOptions(t *testing.T) {
	h, cookie := newAuthedAdminHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/reviews/add/", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	bookStart := strings.Index(body, `<span class="field-label">Book</span>`)
	if bookStart < 0 {
		t.Fatal("book field missing")
	}
	bookBlock := body[bookStart:]
	if i := strings.Index(bookBlock, `<span class="field-label">User</span>`); i > 0 {
		bookBlock = bookBlock[:i]
	}
	n := strings.Count(bookBlock, "<option")
	// placeholder "-- Select --" plus DefaultOptionLimit hits
	if n != vent.DefaultOptionLimit+1 {
		t.Fatalf("book <option> count = %d, want %d", n, vent.DefaultOptionLimit+1)
	}
	if strings.Contains(bookBlock, "Needle Title") {
		t.Fatal("book 150 should not appear in the first 100 options")
	}
}

func newTestAdminHandler(t *testing.T) http.Handler {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:fkoptionshttp?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	creds := auth.NewBCryptCredentialGenerator()
	hash, err := creds.Generate("secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.User.Create().
		SetEmail("admin@vent.com").
		SetPasswordHash(hash).
		SetIsStaff(true).
		SetIsSuperuser(true).
		Save(ctx); err != nil {
		t.Fatal(err)
	}
	authorUser, err := client.User.Create().
		SetEmail("author@vent.com").
		SetPasswordHash(hash).
		SetIsStaff(true).
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	a, err := client.Author.Create().SetUserID(authorUser.ID).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 150; i++ {
		title := fmt.Sprintf("Book %03d", i)
		if i == 150 {
			title = "Needle Title"
		}
		if _, err := client.Book.Create().SetTitle(title).SetAuthorID(a.ID).Save(ctx); err != nil {
			t.Fatal(err)
		}
	}

	h, err := admin.NewAdminHandler(admin.AdminConfig{
		Client: client,
		SecretProvider: auth.SecretProviderFunc(func() []byte {
			return []byte("test-secret")
		}),
		CredentialGenerator:     creds,
		CredentialAuthenticator: auth.NewBCryptCredentialAuthenticator(),
		Schemas: admin.SchemaAdmins{
			Book: BookAdmin{DefaultBookAdmin: admin.NewDefaultBookAdmin(client)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func newAuthedAdminHandler(t *testing.T) (http.Handler, *http.Cookie) {
	t.Helper()
	h := newTestAdminHandler(t)
	token, err := auth.NewJwtTokenGenerator(auth.SecretProviderFunc(func() []byte {
		return []byte("test-secret")
	})).Generate(auth.NewClaims(1))
	if err != nil {
		t.Fatal(err)
	}
	return h, &http.Cookie{Name: "vent-auth-token", Value: token}
}
