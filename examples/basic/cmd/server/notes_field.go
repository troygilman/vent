package main

import (
	"context"

	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/templates/gui"
)

// BookNotesField is a virtual admin field declared via CustomFields.
// It reads and writes the sensitive Ent field internal_notes, which is
// intentionally omitted from the default admin surface.
type BookNotesField struct{}

func (BookNotesField) ListCell(_ context.Context, e *ent.Book) string {
	if e.InternalNotes == nil {
		return ""
	}
	return *e.InternalNotes
}

func (BookNotesField) CreateHTML(ctx context.Context) (string, error) {
	return gui.RenderTextFieldHTML(ctx, gui.SchemaEntityTextFieldProps{
		Name:     "notes",
		Label:    "Notes",
		Editable: gui.MustRenderContext(ctx).CanUpdate,
	})
}

func (BookNotesField) UpdateHTML(ctx context.Context, e *ent.Book) (string, error) {
	value := ""
	if e.InternalNotes != nil {
		value = *e.InternalNotes
	}
	return gui.RenderTextFieldHTML(ctx, gui.SchemaEntityTextFieldProps{
		Name:     "notes",
		Label:    "Notes",
		Value:    value,
		Editable: gui.MustRenderContext(ctx).CanUpdate,
	})
}

func (BookNotesField) ApplyCreate(_ context.Context, builder *ent.BookCreate, input admin.BookCreateInput) error {
	if input.Notes != "" {
		builder.SetInternalNotes(input.Notes)
	}
	return nil
}

func (BookNotesField) ApplyUpdate(_ context.Context, builder *ent.BookUpdateOne, input admin.BookUpdateInput) error {
	if input.Notes == nil {
		return nil
	}
	if *input.Notes == "" {
		builder.ClearInternalNotes()
		return nil
	}
	builder.SetInternalNotes(*input.Notes)
	return nil
}
