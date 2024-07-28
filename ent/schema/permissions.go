package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Permissions holds the schema definition for the Permissions entity.
type Permissions struct {
	ent.Schema
}

// Fields of the Permissions.
func (Permissions) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id,omitempty"`),
		field.String("name").
			Unique(),
		field.String("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Permissions.
func (Permissions) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Roles.Type).
			Ref("permissions"),
	}
}
