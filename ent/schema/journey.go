package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Journey holds the schema definition for the Journey entity.
type Journey struct {
	ent.Schema
}

// Fields of the Journey.
func (Journey) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("org_id", uuid.UUID{}), // RLS column
		field.UUID("product_id", uuid.UUID{}),
		field.UUID("feature_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.Text("description").
			Optional(),

		// Journey matching configuration
		field.JSON("entry_conditions", []any{}).
			Optional(),
		field.JSON("exit_conditions", []any{}).
			Optional(),
		field.JSON("steps", []any{}).
			Optional(),
		field.Int("timeout_minutes").
			Default(30),

		// Analytics (denormalized for quick access)
		field.Int64("total_sessions").
			Default(0),
		field.Int64("converted_sessions").
			Default(0),
		field.Float("conversion_rate").
			Default(0),
		field.Int64("avg_duration_ms").
			Default(0),

		// Status
		field.Bool("is_active").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Journey.
func (Journey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("journeys").
			Field("product_id").
			Unique().
			Required(),
		edge.From("feature", Feature.Type).
			Ref("journeys").
			Field("feature_id").
			Unique(),
		edge.To("events", Event.Type),
		edge.To("sessions", Session.Type),
	}
}

// Indexes of the Journey.
func (Journey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "product_id"),
		index.Fields("org_id", "is_active"),
		index.Fields("org_id", "feature_id"),
		index.Fields("product_id", "name").Unique(),
	}
}
