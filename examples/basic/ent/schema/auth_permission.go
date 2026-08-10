package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/troygilman/vent"
)

type AuthPermission struct {
	ent.Schema
}

func (AuthPermission) Fields() []ent.Field {
	return nil
}
func (AuthPermission) Edges() []ent.Edge {
	return nil
}
func (AuthPermission) Annotations() []schema.Annotation {
	return nil
}
func (AuthPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.AuthPermissionMixin{
			GroupSchemaType: AuthGroup.Type,
		},
	}
}
