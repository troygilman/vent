package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/troygilman/vent"
)

// Permission uses PermissionMixin defaults: DisableCreate, DisableDelete, and
// ReadOnlyFields for "name" (permissions are managed by the migrator).
// FilterableColumns is omitted so the list page has no filters (this schema
// annotation replaces the mixin, which would otherwise filter on name). That
// path is how we test chrome-first loading without a filter form.
type Permission struct {
	ent.Schema
}

func (Permission) Fields() []ent.Field {
	return nil
}

func (Permission) Edges() []ent.Edge {
	return nil
}

func (Permission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			SingularDisplayName: "Permission",
			PluralDisplayName:   "Permissions",
			DisableCreate:       true,
			DisableDelete:       true,
			ReadOnlyFields:      []string{"name"},
			TableColumns:        []string{"name", "groups"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{"name", "groups"},
			}},
		},
	}
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.PermissionMixin{
			GroupSchemaType: PermissionGroup.Type,
		},
	}
}
