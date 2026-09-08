package vent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/templates/gui"

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

	userOpts := decodeOptions(t, get("/admin/reviews/options/user/?q=alice", true))
	if !hasOptionValue(userOpts, alice.ID) {
		t.Fatalf("user options = %#v, want alice id %d", userOpts, alice.ID)
	}

	bookOpts := decodeOptions(t, get("/admin/reviews/options/book/?q=Zebra", true))
	if !hasOptionValue(bookOpts, book.ID) {
		t.Fatalf("book options = %#v, want book id %d", bookOpts, book.ID)
	}
	for _, opt := range bookOpts {
		if strings.Contains(opt.Label, "alice") {
			t.Fatalf("book options included a user: %#v", bookOpts)
		}
	}

	userAgain := decodeOptions(t, get("/admin/reviews/options/user/?q=alice", true))
	if !hasOptionValue(userAgain, alice.ID) {
		t.Fatalf("second user options = %#v, want alice after book request", userAgain)
	}

	groups := get("/admin/users/options/groups/", true)
	if groups.Code != http.StatusOK {
		t.Fatalf("user groups options status = %d body = %q, want 200", groups.Code, groups.Body.String())
	}
}

func decodeOptions(t *testing.T, rec *httptest.ResponseRecorder) []gui.SelectOption {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("options status = %d body = %q, want 200", rec.Code, rec.Body.String())
	}
	var options []gui.SelectOption
	if err := json.Unmarshal(rec.Body.Bytes(), &options); err != nil {
		t.Fatalf("decode options: %v body = %q", err, rec.Body.String())
	}
	return options
}

func hasOptionValue(options []gui.SelectOption, id int) bool {
	for _, opt := range options {
		if opt.Value == id {
			return true
		}
	}
	return false
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
