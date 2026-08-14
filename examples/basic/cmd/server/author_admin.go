package main

import (
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
	return fmt.Sprintf("%d", e.ID)
}
