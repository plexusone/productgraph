package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Capability holds the schema definition for the Capability entity.
// Capabilities represent major functional areas within a Product.
type Capability struct {
	ent.Schema
}

// Fields of the Capability.
func (Capability) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("org_id", uuid.UUID{}), // RLS column
		field.UUID("product_id", uuid.UUID{}),
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.Text("description").
			Optional(),
		field.Int("display_order").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Capability.
func (Capability) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("capabilities").
			Field("product_id").
			Unique().
			Required(),
		edge.To("features", Feature.Type),
	}
}

// Indexes of the Capability.
func (Capability) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "product_id"),
		index.Fields("product_id", "name").Unique(),
	}
}
