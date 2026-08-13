package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/troygilman/vent"
)

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
	return nil
}
func (PermissionGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.PermissionGroupMixin{
			UserSchemaType:       User.Type,
			PermissionSchemaType: Permission.Type,
		},
	}
}
