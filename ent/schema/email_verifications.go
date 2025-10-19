package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// EmailVerifications holds the schema definition for the EmailVerifications entity.
type EmailVerifications struct {
	ent.Schema
}

// Fields of the EmailVerifications.
func (EmailVerifications) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}).
			Comment("User requesting verification"),
		field.String("email").
			NotEmpty(),
		field.String("token").
			NotEmpty().
			Unique(),
		field.Time("expires_at").
			Comment("When this token expires"),
		field.Bool("is_used").
			Default(false),
		field.Time("used_at").
			Optional().
			Nillable(),
		field.String("ip_address").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the EmailVerifications.
func (EmailVerifications) Edges() []ent.Edge {
	return nil
}
