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
		field.String("code").
			Unique().
			NotEmpty().
			Comment("Unique code identifier for the role"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Bool("is_system").
			Default(false).
			Comment("System roles cannot be deleted or modified via API"),
		field.Bool("is_default").
			Default(false).
			Comment("Default role assigned to new users on signup"),
		field.Int("max_users").
			Optional().
			Nillable().
			Comment("Maximum users allowed for this role (null = unlimited)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Roles.
func (Roles) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user_roles", UserRoles.Type).
			Ref("role"),
		edge.From("role_permissions", RolePermissions.Type).
			Ref("role"),
	}
}
