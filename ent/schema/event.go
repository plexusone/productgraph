package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("org_id", uuid.UUID{}), // RLS column
		field.UUID("project_id", uuid.UUID{}),
		field.String("event_id").
			NotEmpty().
			MaxLen(255),
		field.String("session_id").
			NotEmpty().
			MaxLen(255),
		field.String("user_id").
			Optional().
			MaxLen(255),

		// Classification (maps to OTel event.* namespace)
		field.Enum("event_type").
			Values(
				"page_view", "page_leave",
				"ui_click", "ui_input", "ui_scroll", "ui_focus", "ui_blur", "ui_submit",
				"state_change",
				"api_request", "api_response",
				"journey_step",
				"error", "performance", "custom",
			),
		field.String("event_name").
			Optional().
			MaxLen(255),
		field.Time("timestamp"),
		field.Int64("sequence").
			Default(0),

		// Page context
		field.String("page_path").
			Optional().
			MaxLen(1024),
		field.String("page_title").
			Optional().
			MaxLen(255),
		field.String("page_url").
			Optional().
			MaxLen(2048),
		field.String("page_referrer").
			Optional().
			MaxLen(2048),

		// UI context
		field.String("ui_component_name").
			Optional().
			MaxLen(255),
		field.String("ui_component_path").
			Optional().
			MaxLen(1024),
		field.String("ui_component_type").
			Optional().
			MaxLen(100),
		field.String("ui_action").
			Optional().
			MaxLen(100),
		field.String("ui_element").
			Optional().
			MaxLen(255),
		field.String("ui_element_text").
			Optional().
			MaxLen(500),
		field.String("ui_viewport").
			Optional().
			MaxLen(50),
		field.Float("ui_scroll_position").
			Optional(),

		// State tracking
		field.String("ui_state_key").
			Optional().
			MaxLen(255),
		field.Text("ui_state_before").
			Optional(),
		field.Text("ui_state_after").
			Optional(),
		field.String("ui_state_change_type").
			Optional().
			MaxLen(50),

		// Journey context
		field.UUID("journey_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("journey_step_id").
			Optional().
			MaxLen(255),
		field.String("journey_step_name").
			Optional().
			MaxLen(255),
		field.String("conversion_status").
			Optional().
			MaxLen(50),

		// API tracking
		field.String("api_method").
			Optional().
			MaxLen(10),
		field.String("api_path").
			Optional().
			MaxLen(1024),
		field.Int("api_status_code").
			Optional(),
		field.Int64("api_duration_ms").
			Optional(),

		// Error tracking
		field.String("error_type").
			Optional().
			MaxLen(255),
		field.Text("error_message").
			Optional(),
		field.Text("error_stack").
			Optional(),
		field.String("error_component").
			Optional().
			MaxLen(255),

		// Performance (Web Vitals)
		field.Float("performance_lcp_ms").
			Optional(),
		field.Float("performance_fid_ms").
			Optional(),
		field.Float("performance_cls").
			Optional(),
		field.Float("performance_ttfb_ms").
			Optional(),

		// Snapshot
		field.String("snapshot_url").
			Optional().
			MaxLen(1024),
		field.String("snapshot_viewport").
			Optional().
			MaxLen(50),

		// Duration
		field.Int64("duration_ms").
			Optional(),

		// Custom metadata
		field.JSON("metadata", map[string]any{}).
			Optional(),

		// Timestamps
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("events").
			Field("project_id").
			Unique().
			Required(),
		edge.From("journey", Journey.Type).
			Ref("events").
			Field("journey_id").
			Unique(),
	}
}

// Indexes of the Event.
func (Event) Indexes() []ent.Index {
	return []ent.Index{
		// Primary query patterns
		index.Fields("org_id", "project_id", "timestamp"),
		index.Fields("org_id", "session_id"),
		index.Fields("org_id", "user_id"),
		index.Fields("org_id", "journey_id"),
		index.Fields("org_id", "event_type", "timestamp"),
		// Unique constraint for event_id within project
		index.Fields("project_id", "event_id").Unique(),
	}
}
