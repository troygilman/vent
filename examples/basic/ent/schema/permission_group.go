package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/troygilman/vent"
)

// PermissionGroup is the RBAC grouping schema from Vent's auth mixin.
type PermissionGroup struct {
	ent.Schema
}

func (PermissionGroup) Fields() []ent.Field {
	return nil
}

func (PermissionGroup) Edges() []ent.Edge {
	return nil
}

func (PermissionGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			SingularDisplayName: "Permission Group",
			PluralDisplayName:   "Permission Groups",
			RouteName:           "permission-groups",
			TableColumns:        []string{"name"},
			FilterableColumns:   []string{"name"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{"name", "permissions"},
			}},
		},
	}
}

func (PermissionGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.PermissionGroupMixin{
			UserSchemaType:       User.Type,
			PermissionSchemaType: Permission.Type,
		},
	}
}
