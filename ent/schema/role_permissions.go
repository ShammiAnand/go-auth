package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RolePermissions holds the schema definition for the RolePermissions entity (join table).
type RolePermissions struct {
	ent.Schema
}

// Fields of the RolePermissions.
func (RolePermissions) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.Int("role_id"),
		field.Int("permission_id"),
		field.Time("assigned_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the RolePermissions.
func (RolePermissions) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", Roles.Type).
			Unique().
			Required().
			Field("role_id"),
		edge.To("permission", Permissions.Type).
			Unique().
			Required().
			Field("permission_id"),
	}
}

// Indexes of the RolePermissions.
func (RolePermissions) Indexes() []ent.Index {
	return []ent.Index{
		// Unique constraint on role_id + permission_id
		index.Fields("role_id", "permission_id").
			Unique(),
	}
}
