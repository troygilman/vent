package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/author"
	"github.com/troygilman/vent/examples/basic/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

func TestAuthorAdminNameUsesEagerLoadedEmail(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:author_admin?mode=memory&cache=shared&_fk=1")
	user := client.User.Create().
		SetEmail("ada@vent.com").
		SaveX(ctx)
	created := client.Author.Create().
		SetUser(user).
		SaveX(ctx)

	a := AuthorAdmin{DefaultAuthorAdmin: admin.NewDefaultAuthorAdmin(client)}

	unloaded := client.Author.GetX(ctx, created.ID)
	if got := a.Name(unloaded); got != fmt.Sprintf("%d", created.ID) {
		t.Fatalf("Name() without user edge = %q, want id", got)
	}

	loaded := a.EagerLoadQuery(client.Author.Query().Where(author.IDEQ(created.ID))).OnlyX(ctx)
	if got := a.Name(loaded); got != "ada@vent.com" {
		t.Fatalf("Name() after EagerLoadQuery = %q, want %q", got, "ada@vent.com")
	}
}
