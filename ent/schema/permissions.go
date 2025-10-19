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
		field.String("code").
			Unique().
			NotEmpty().
			Comment("Unique code identifier (e.g., users.read, rbac.write)"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("resource").
			Optional().
			Comment("Resource this permission applies to (e.g., users, roles)"),
		field.String("action").
			Optional().
			Comment("Action this permission allows (e.g., read, write, delete)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Permissions.
func (Permissions) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("role_permissions", RolePermissions.Type).
			Ref("permission"),
	}
}
