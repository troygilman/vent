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

func TestFKOptionsUnauthenticatedRedirects(t *testing.T) {
	h := newTestAdminHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/reviews/options/book/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "login") {
		t.Fatalf("Location = %q, want login redirect", loc)
	}
}

func TestFKOptionsUnknownEdgeBadRequest(t *testing.T) {
	h, cookie := newAuthedAdminHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/reviews/options/not-an-edge/", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s, want 400", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unknown edge") {
		t.Fatalf("body = %q, want unknown edge", rec.Body.String())
	}
}

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
	if strings.Count(body, `id="entity-book-options"`) != 1 {
		t.Fatal("expected one book option list")
	}
	bookBlock := body[strings.Index(body, `id="entity-book-options"`):]
	if i := strings.Index(bookBlock, `id="entity-user-options"`); i > 0 {
		bookBlock = bookBlock[:i]
	}
	n := strings.Count(bookBlock, `role="option"`)
	if n != vent.DefaultOptionLimit {
		t.Fatalf("book options on add form = %d, want %d (got unbounded dump?)", n, vent.DefaultOptionLimit)
	}
	if strings.Contains(body, "Needle Title") {
		t.Fatal("book 150 should not appear in the first 100 options")
	}
}

func TestReviewBookOptionsSearchAndSelectedUnion(t *testing.T) {
	h, cookie := newAuthedAdminHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/reviews/options/book/?q=Needle", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Needle Title") {
		t.Fatalf("search missed Needle Title: %s", body)
	}
	if strings.Count(body, `role="option"`) > vent.DefaultOptionLimit {
		t.Fatal("search results exceeded DefaultOptionLimit")
	}
}

func newTestAdminHandler(t *testing.T) http.Handler {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:fkoptions?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	creds := auth.NewBCryptCredentialGenerator()
	hash, err := creds.Generate("secret")
	if err != nil {
		t.Fatal(err)
	}
	adminUser, err := client.User.Create().
		SetEmail("admin@vent.com").
		SetPasswordHash(hash).
		SetIsStaff(true).
		SetIsSuperuser(true).
		Save(ctx)
	if err != nil {
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
	_ = adminUser

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
