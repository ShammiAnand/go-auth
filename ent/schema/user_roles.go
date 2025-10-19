package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserRoles holds the schema definition for the UserRoles entity (join table).
type UserRoles struct {
	ent.Schema
}

// Fields of the UserRoles.
func (UserRoles) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.Int("role_id"),
		field.UUID("assigned_by", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("User who assigned this role"),
		field.Time("assigned_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the UserRoles.
func (UserRoles) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", Users.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("role", Roles.Type).
			Unique().
			Required().
			Field("role_id"),
	}
}

// Indexes of the UserRoles.
func (UserRoles) Indexes() []ent.Index {
	return []ent.Index{
		// Unique constraint on user_id + role_id
		index.Fields("user_id", "role_id").
			Unique(),
	}
}
