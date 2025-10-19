package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"
)

// Users holds the schema definition for the Users entity.
type Users struct {
	ent.Schema
}

// Fields of the Users.
func (Users) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("email").NotEmpty().Unique(),
		field.String("password_hash").NotEmpty(),
		field.String("first_name").NotEmpty(),
		field.String("last_name").NotEmpty(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("last_login").Optional(),
		field.Bool("is_active").
			Default(true),
		field.Bool("email_verified").
			Default(false),

		// NOTE: below field are for implementing passwoord reset and email verification
		field.String("verification_token").
			Optional().
			Nillable(),
		field.Time("verification_token_expiry").
			Optional().
			Nillable(),
		field.String("password_reset_token").
			Optional().
			Nillable(),
		field.Time("password_reset_token_expiry").
			Optional().
			Nillable(),

		// NOTE: for future needs
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
	}
}

// Edges of the Users.
func (Users) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user_roles", UserRoles.Type).
			Ref("user"),
	}
}
