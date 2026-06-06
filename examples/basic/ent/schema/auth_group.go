package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/troygilman/vent"
)

type AuthGroup struct {
	ent.Schema
}

func (AuthGroup) Fields() []ent.Field {
	return nil
}
func (AuthGroup) Edges() []ent.Edge {
	return nil
}
func (AuthGroup) Annotations() []schema.Annotation {
	return nil
}
func (AuthGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.AuthGroupMixin{
			UserSchemaType:       AuthUser.Type,
			PermissionSchemaType: AuthPermission.Type,
		},
	}
}
