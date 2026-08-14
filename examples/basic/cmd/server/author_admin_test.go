package main

import (
	"context"
	"testing"

	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

func TestAuthorAdminNameUsesEmail(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	user := client.User.Create().
		SetEmail("ada@vent.com").
		SaveX(ctx)
	author := client.Author.Create().
		SetUser(user).
		SaveX(ctx)

	loaded := client.Author.GetX(ctx, author.ID)
	if loaded.Edges.User != nil {
		t.Fatal("expected author loaded without user edge")
	}

	a := AuthorAdmin{DefaultAuthorAdmin: admin.NewDefaultAuthorAdmin(client)}
	if got := a.Name(loaded); got != "ada@vent.com" {
		t.Fatalf("Name() = %q, want %q", got, "ada@vent.com")
	}
}
