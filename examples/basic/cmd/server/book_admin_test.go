package main

import (
	"context"
	"testing"

	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/book"
	"github.com/troygilman/vent/examples/basic/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

func TestBookAdminEagerLoadQueryNestsAuthorUser(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:book_admin?mode=memory&cache=shared&_fk=1")
	user := client.User.Create().
		SetEmail("ada@vent.com").
		SaveX(ctx)
	author := client.Author.Create().
		SetUser(user).
		SaveX(ctx)
	created := client.Book.Create().
		SetTitle("Analytical Engines").
		SetAuthor(author).
		SaveX(ctx)

	defaultLoaded := admin.NewDefaultBookAdmin(client).
		EagerLoadQuery(client.Book.Query().Where(book.IDEQ(created.ID))).
		OnlyX(ctx)
	if defaultLoaded.Edges.Author == nil {
		t.Fatal("default EagerLoadQuery did not load author")
	}
	if defaultLoaded.Edges.Author.Edges.User != nil {
		t.Fatal("default EagerLoadQuery nested user; want a one-level WithAuthor only")
	}

	custom := BookAdmin{DefaultBookAdmin: admin.NewDefaultBookAdmin(client)}
	loaded := custom.EagerLoadQuery(client.Book.Query().Where(book.IDEQ(created.ID))).
		OnlyX(ctx)
	if loaded.Edges.Author == nil || loaded.Edges.Author.Edges.User == nil {
		t.Fatal("BookAdmin.EagerLoadQuery did not nest author.user")
	}
	if got := loaded.Edges.Author.Edges.User.Email; got != "ada@vent.com" {
		t.Fatalf("nested user email = %q, want ada@vent.com", got)
	}

	a := AuthorAdmin{DefaultAuthorAdmin: admin.NewDefaultAuthorAdmin(client)}
	if got := a.Name(loaded.Edges.Author); got != "ada@vent.com" {
		t.Fatalf("Name() = %q, want %q", got, "ada@vent.com")
	}
}
