package main

import (
	ent "github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
)

// BookAdmin customizes the book admin surface.
// Embed DefaultBookAdmin and override only what you need.
type BookAdmin struct {
	admin.DefaultBookAdmin
}

func (BookAdmin) Name(e *ent.Book) string {
	return e.Title
}

// FieldNotes supplies the required custom field declared on the Book schema.
func (a BookAdmin) FieldNotes() admin.BookField {
	return BookNotesField{}
}
