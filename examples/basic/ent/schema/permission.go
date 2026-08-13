package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/troygilman/vent"
)

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
	return nil
}
func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.PermissionMixin{
			GroupSchemaType: PermissionGroup.Type,
		},
	}
}
