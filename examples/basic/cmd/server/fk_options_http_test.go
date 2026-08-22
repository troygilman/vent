package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	if strings.Contains(bookBlock, `type="search"`) {
		t.Fatal("FK widget should be a single combobox input, not type=search plus an always-open list")
	}
	if !strings.Contains(bookBlock, `role="combobox"`) {
		t.Fatal("expected combobox role on the FK input")
	}
	if strings.Contains(bookBlock, `data-on:mousedown__prevent `) || strings.Contains(bookBlock, `data-on:mousedown__prevent>`) {
		t.Fatal("Datastar data-on requires a value; mousedown__prevent must not be a boolean attribute")
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

func TestReviewBookOptionsSearchDoesNotInjectSelected(t *testing.T) {
	h, cookie := newAuthedAdminHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		`/admin/reviews/options/book/?q=Needle&datastar=`+url.QueryEscape(`{"entity":{"book":"1"}}`),
		nil,
	)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Needle Title") {
		t.Fatalf("search missed Needle Title: %s", body)
	}
	if strings.Count(body, `role="option"`) != 1 {
		t.Fatalf("search options = %d, want hits only", strings.Count(body, `role="option"`))
	}
	if strings.Contains(body, `role="option"`) && strings.Contains(body[strings.Index(body, `role="option"`):], "Book 001") {
		t.Fatal("selected Book 001 must not be injected into Needle search hits")
	}
	if !strings.Contains(body, `class="fk-chip"`) || !strings.Contains(body, "Book 001") {
		t.Fatalf("selected book should remain a chip: %s", body)
	}
}

func TestReviewUserOptionsSearchDoesNotUseBookEdge(t *testing.T) {
	h, cookie := newAuthedAdminHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/reviews/options/user/?q=author@vent.com", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, `id="entity-book-options"`) {
		t.Fatalf("user options URL rendered book list: %s", body)
	}
	if !strings.Contains(body, `id="entity-user-options"`) {
		t.Fatalf("expected entity-user-options, got %s", body)
	}
	if strings.Count(body, `role="option"`) < 1 {
		t.Fatalf("user email search returned no options: %s", body)
	}
}

func TestReviewOptionsEdgesDoNotStickAcrossRequests(t *testing.T) {
	h, cookie := newAuthedAdminHandler(t)
	get := func(path string) string {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(cookie)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", path, rec.Code, rec.Body.String())
		}
		return rec.Body.String()
	}
	book := get("/admin/reviews/options/book/?q=Needle")
	if !strings.Contains(book, `id="entity-book-options"`) || !strings.Contains(book, "Needle Title") {
		t.Fatalf("book search = %s", book)
	}
	user := get("/admin/reviews/options/user/?q=author@vent.com")
	if strings.Contains(user, `id="entity-book-options"`) {
		t.Fatalf("user search reused book edge: %s", user)
	}
	if !strings.Contains(user, `id="entity-user-options"`) {
		t.Fatalf("user search = %s", user)
	}
}

func TestBookAuthorOptionsSearchViaUserEmail(t *testing.T) {
	h, cookie := newAuthedAdminHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/books/options/author/?q=author@vent.com", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `role="option"`) {
		t.Fatalf("author search returned no options: %s", body)
	}
	if strings.Contains(body, "No matches") {
		t.Fatalf("author search should filter via user email, got %s", body)
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
