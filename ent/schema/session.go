package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("org_id", uuid.UUID{}), // RLS column
		field.UUID("project_id", uuid.UUID{}),
		field.String("session_id").
			NotEmpty().
			MaxLen(255),
		field.String("user_id").
			Optional().
			MaxLen(255),

		// Session timing
		field.Time("started_at"),
		field.Time("ended_at").
			Optional().
			Nillable(),
		field.Int64("duration_ms").
			Default(0),

		// Navigation
		field.String("entry_page").
			Optional().
			MaxLen(1024),
		field.String("exit_page").
			Optional().
			MaxLen(1024),
		field.Int("page_count").
			Default(0),
		field.Int("event_count").
			Default(0),

		// Journey tracking
		field.UUID("journey_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("last_journey_step_id").
			Optional().
			MaxLen(255),
		field.String("conversion_status").
			Optional().
			MaxLen(50),

		// Device info
		field.String("user_agent").
			Optional().
			MaxLen(1024),
		field.String("device_type").
			Optional().
			MaxLen(50),
		field.String("browser").
			Optional().
			MaxLen(100),
		field.String("os").
			Optional().
			MaxLen(100),

		// Geo (optional, from IP)
		field.String("country").
			Optional().
			MaxLen(2),
		field.String("region").
			Optional().
			MaxLen(100),
		field.String("city").
			Optional().
			MaxLen(100),

		// Timestamps
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("sessions").
			Field("project_id").
			Unique().
			Required(),
		edge.From("journey", Journey.Type).
			Ref("sessions").
			Field("journey_id").
			Unique(),
	}
}

// Indexes of the Session.
func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "project_id", "started_at"),
		index.Fields("org_id", "session_id"),
		index.Fields("org_id", "user_id"),
		index.Fields("org_id", "journey_id"),
		index.Fields("org_id", "conversion_status"),
		// Unique constraint for session_id within project
		index.Fields("project_id", "session_id").Unique(),
	}
}
