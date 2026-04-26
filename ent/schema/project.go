package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("org_id", uuid.UUID{}),
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.String("slug").
			NotEmpty().
			MaxLen(100),
		field.String("api_key").
			Unique().
			NotEmpty().
			MaxLen(64).
			Sensitive(),
		field.JSON("settings", map[string]any{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("projects").
			Field("org_id").
			Unique().
			Required(),
		edge.To("events", Event.Type),
		edge.To("sessions", Session.Type),
		edge.To("journeys", Journey.Type),
	}
}

// Indexes of the Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "slug").Unique(),
		index.Fields("api_key"),
	}
}
