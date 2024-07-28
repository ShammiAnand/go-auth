package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Roles holds the schema definition for the Roles entity.
type Roles struct {
	ent.Schema
}

// Fields of the Roles.
func (Roles) Fields() []ent.Field {
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

// Edges of the Roles.
func (Roles) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", Users.Type).
			Ref("roles"),
		edge.To("permissions", Permissions.Type),
	}
}
