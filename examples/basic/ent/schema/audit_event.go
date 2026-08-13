package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/troygilman/vent"
)

// AuditEvent demonstrates a fully read-only admin schema (no create/update/delete).
type AuditEvent struct {
	ent.Schema
}

func (AuditEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("action").NotEmpty(),
		field.String("detail").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (AuditEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		vent.VentSchemaAnnotation{
			ReadOnly:            true,
			SingularDisplayName: "Audit Event",
			PluralDisplayName:   "Audit Events",
			RouteName:           "audit-events",
			ReadOnlyFields:      []string{"action", "detail", "created_at"},
			TableColumns:        []string{"action", "detail", "created_at"},
			FieldSets: []vent.FieldSet{{
				Fields: []string{"action", "detail", "created_at"},
			}},
		},
	}
}
