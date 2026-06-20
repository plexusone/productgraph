package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
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

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("products").
			Field("org_id").
			Unique().
			Required(),
		edge.To("capabilities", Capability.Type),
		edge.To("events", Event.Type),
		edge.To("sessions", Session.Type),
		edge.To("journeys", Journey.Type),
	}
}

// Indexes of the Product.
func (Product) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "slug").Unique(),
		index.Fields("api_key"),
	}
}
