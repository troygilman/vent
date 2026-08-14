package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/troygilman/vent"
)

// Permission uses PermissionMixin defaults: DisableCreate, DisableDelete, and
// ReadOnlyFields for "name" (permissions are managed by the migrator).
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
			FilterableColumns:   []string{"name"},
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
