package main

import (
	"context"
	"fmt"

	ent "github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
)

// AuthorAdmin labels authors by their linked user email.
type AuthorAdmin struct {
	admin.DefaultAuthorAdmin
}

func (AuthorAdmin) Name(e *ent.Author) string {
	if e.Edges.User != nil {
		return e.Edges.User.Email
	}
	// FK option lists and nested book edges do not always eager-load User.
	u, err := e.QueryUser().Only(context.Background())
	if err != nil {
		return fmt.Sprintf("%d", e.ID)
	}
	return u.Email
}
