package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// EmailLogs holds the schema definition for the EmailLogs entity.
type EmailLogs struct {
	ent.Schema
}

// Fields of the EmailLogs.
func (EmailLogs) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("User this email was sent to"),
		field.String("recipient").
			NotEmpty().
			Comment("Email address of recipient"),
		field.String("email_type").
			NotEmpty().
			Comment("Type of email (verification, password_reset, welcome)"),
		field.String("subject").
			Optional(),
		field.String("status").
			Default("sent").
			Comment("Status: sent, delivered, failed, bounced"),
		field.String("provider").
			Default("mailhog").
			Comment("Email provider used (ses, mailhog)"),
		field.String("provider_message_id").
			Optional().
			Comment("Message ID from email provider"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.String("error_message").
			Optional().
			Comment("Error message if delivery failed"),
		field.Time("sent_at").
			Default(time.Now).
			Immutable(),
		field.Time("delivered_at").
			Optional().
			Nillable(),
	}
}

// Edges of the EmailLogs.
func (EmailLogs) Edges() []ent.Edge {
	return nil
}
