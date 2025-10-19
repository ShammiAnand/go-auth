package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// AuditLogs holds the schema definition for the AuditLogs entity.
type AuditLogs struct {
	ent.Schema
}

// Fields of the AuditLogs.
func (AuditLogs) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("actor_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("User who performed the action"),
		field.String("action_type").
			NotEmpty().
			Comment("Type of action (e.g., role.create, permission.assign)"),
		field.String("resource_type").
			NotEmpty().
			Comment("Type of resource affected (e.g., role, user_role)"),
		field.String("resource_id").
			Optional().
			Comment("ID of the affected resource"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Additional context about the action"),
		field.JSON("changes", map[string]interface{}{}).
			Optional().
			Comment("Before/after values for updates"),
		field.String("ip_address").
			Optional(),
		field.String("user_agent").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the AuditLogs.
func (AuditLogs) Edges() []ent.Edge {
	return nil
}
