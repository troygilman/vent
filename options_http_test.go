package vent_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"

	_ "github.com/mattn/go-sqlite3"
)

func TestFKOptionsHTTPRoute(t *testing.T) {
	ctx := context.Background()
	client, err := ent.Open("sqlite3", "file:options_http?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("schema create: %v", err)
	}

	secret := auth.SecretProviderFunc(func() []byte { return []byte("secret") })
	adminUser, err := client.User.Create().
		SetEmail("admin@vent.com").
		SetIsStaff(true).
		SetIsSuperuser(true).
		SetIsActive(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	alice, err := client.User.Create().
		SetEmail("alice@vent.com").
		SetIsStaff(true).
		SetIsActive(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create alice: %v", err)
	}
	author, err := client.Author.Create().
		SetUserID(alice.ID).
		SetActive(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create author: %v", err)
	}
	book, err := client.Book.Create().
		SetTitle("Zebra Unique Title").
		SetAuthorID(author.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	handler, err := admin.NewAdminHandler(admin.AdminConfig{
		Client:                  client,
		SecretProvider:          secret,
		CredentialAuthenticator: auth.NewBCryptCredentialAuthenticator(),
		CredentialGenerator:     auth.NewBCryptCredentialGenerator(),
		Schemas: admin.SchemaAdmins{
			Book: testBookAdmin{DefaultBookAdmin: admin.NewDefaultBookAdmin(client)},
		},
	})
	if err != nil {
		t.Fatalf("NewAdminHandler: %v", err)
	}

	token, err := auth.NewJwtTokenGenerator(secret).Generate(auth.NewClaims(adminUser.ID))
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	get := func(path string, authed bool) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if authed {
			req.AddCookie(&http.Cookie{Name: "vent-auth-token", Value: token})
		}
		handler.ServeHTTP(rec, req)
		return rec
	}

	unauth := get("/admin/reviews/options/user/", false)
	if unauth.Code != http.StatusSeeOther {
		t.Fatalf("unauthenticated status = %d, want 303", unauth.Code)
	}
	if loc := unauth.Header().Get("Location"); !strings.Contains(loc, "/admin/login/") {
		t.Fatalf("unauthenticated Location = %q", loc)
	}

	unknown := get("/admin/reviews/options/not-an-edge/", true)
	if unknown.Code != http.StatusNotFound && unknown.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unknown edge status = %d body = %q, want 404 or 405", unknown.Code, unknown.Body.String())
	}

	userHTML := getHTML(t, get("/admin/reviews/options/user/?q=alice", true))
	if !strings.Contains(userHTML, `id="fk-field-user"`) {
		t.Fatalf("user field html missing id, got %s", userHTML)
	}
	if !strings.Contains(userHTML, fmt.Sprintf(`value="%d"`, alice.ID)) {
		t.Fatalf("user field html missing alice id %d, got %s", alice.ID, userHTML)
	}

	bookHTML := getHTML(t, get("/admin/reviews/options/book/?q=Zebra", true))
	if !strings.Contains(bookHTML, `id="fk-field-book"`) {
		t.Fatalf("book field html missing id, got %s", bookHTML)
	}
	if !strings.Contains(bookHTML, fmt.Sprintf(`value="%d"`, book.ID)) {
		t.Fatalf("book field html missing book id %d, got %s", book.ID, bookHTML)
	}
	if strings.Contains(bookHTML, `id="fk-field-user"`) {
		t.Fatalf("book field html rendered the user field: %s", bookHTML)
	}

	userAgain := getHTML(t, get("/admin/reviews/options/user/?q=alice", true))
	if !strings.Contains(userAgain, fmt.Sprintf(`value="%d"`, alice.ID)) {
		t.Fatalf("second user field html missing alice after book request")
	}

	groups := get("/admin/users/options/groups/", true)
	if groups.Code != http.StatusOK {
		t.Fatalf("user groups options status = %d body = %q, want 200", groups.Code, groups.Body.String())
	}
	if ct := groups.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("groups Content-Type = %q, want text/html", ct)
	}

	ds := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/reviews/options/user/?q=alice", nil)
	req.AddCookie(&http.Cookie{Name: "vent-auth-token", Value: token})
	req.Header.Set("Datastar-Request", "true")
	handler.ServeHTTP(ds, req)
	if ds.Code != http.StatusOK {
		t.Fatalf("datastar options status = %d, want 200", ds.Code)
	}
	if !strings.Contains(ds.Body.String(), fmt.Sprintf(`value="%d"`, alice.ID)) {
		t.Fatalf("datastar patch missing alice, got %s", ds.Body.String())
	}
}

func getHTML(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("options status = %d body = %q, want 200", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	return rec.Body.String()
}

type testBookAdmin struct {
	admin.DefaultBookAdmin
}

func (testBookAdmin) FieldNotes() admin.BookField {
	return stubBookNotesField{}
}

type stubBookNotesField struct{}

func (stubBookNotesField) ListCell(context.Context, *ent.Book) string { return "" }
func (stubBookNotesField) CreateHTML(context.Context) (string, error) { return "", nil }
func (stubBookNotesField) UpdateHTML(context.Context, *ent.Book) (string, error) {
	return "", nil
}
func (stubBookNotesField) ApplyCreate(context.Context, *ent.BookCreate, admin.BookCreateInput) error {
	return nil
}
func (stubBookNotesField) ApplyUpdate(context.Context, *ent.BookUpdateOne, admin.BookUpdateInput) error {
	return nil
}
