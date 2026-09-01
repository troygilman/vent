package admin

import (
	"context"
	"fmt"
	"testing"

	"github.com/troygilman/vent"
	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

func TestReviewBookLoadOptionsLimitAndSearch(t *testing.T) {
	client, ctx := newFKOptionTestClient(t)
	field := NewReviewBookField(client)

	empty, err := field.LoadOptions(ctx, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != vent.DefaultOptionLimit {
		t.Fatalf("empty search options = %d, want %d", len(empty), vent.DefaultOptionLimit)
	}
	for _, opt := range empty {
		if opt.Label == "Needle Title" {
			t.Fatal("book 150 should not appear in the first 100 by ID")
		}
	}

	hits, err := field.LoadOptions(ctx, "Needle", []int{empty[0].Value})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("search hits = %d, want 1", len(hits))
	}
	if hits[0].Label != "Needle Title" {
		t.Fatalf("search label = %q, want Needle Title", hits[0].Label)
	}
	if hits[0].Selected {
		t.Fatal("selected IDs must not be injected into a non-empty search")
	}
}

func TestReviewBookLoadOptionsEmptySearchUnionsSelected(t *testing.T) {
	client, ctx := newFKOptionTestClient(t)
	field := NewReviewBookField(client)

	needle, err := client.Book.Query().All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var needleID int
	for _, b := range needle {
		if b.Title == "Needle Title" {
			needleID = b.ID
			break
		}
	}
	if needleID == 0 {
		t.Fatal("needle book missing")
	}

	opts, err := field.LoadOptions(ctx, "", []int{needleID})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != vent.DefaultOptionLimit+1 {
		t.Fatalf("empty+selected = %d, want %d", len(opts), vent.DefaultOptionLimit+1)
	}
	last := opts[len(opts)-1]
	if last.Value != needleID || !last.Selected || last.Label != "Needle Title" {
		t.Fatalf("union selected = %#v", last)
	}
}

func newFKOptionTestClient(t *testing.T) (*ent.Client, context.Context) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:fkoptions?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("author@vent.com").
		SetPasswordHash("x").
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	a, err := client.Author.Create().SetUserID(u.ID).Save(ctx)
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

	ctx = withAdmin(ctx, SchemaAdmins{
		Author: NewDefaultAuthorAdmin(client),
		Book:   titleBookAdmin{DefaultBookAdmin: NewDefaultBookAdmin(client)},
		Review: NewDefaultReviewAdmin(client),
		User:   NewDefaultUserAdmin(client),
	})
	return client, ctx
}

type titleBookAdmin struct {
	DefaultBookAdmin
}

func (titleBookAdmin) Name(e *ent.Book) string {
	return e.Title
}
