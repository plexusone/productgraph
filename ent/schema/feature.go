package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DocumentationLink represents a link to external documentation.
type DocumentationLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Feature holds the schema definition for the Feature entity.
// Features represent specific functionality within a Capability.
type Feature struct {
	ent.Schema
}

// Fields of the Feature.
func (Feature) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("org_id", uuid.UUID{}), // RLS column
		field.UUID("product_id", uuid.UUID{}),
		field.UUID("capability_id", uuid.UUID{}),
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.Text("description").
			Optional(),
		field.String("admin_path").
			Optional().
			MaxLen(500),
		field.Text("notes").
			Optional(),
		field.JSON("documentation", []DocumentationLink{}).
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

// Edges of the Feature.
func (Feature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("capability", Capability.Type).
			Ref("features").
			Field("capability_id").
			Unique().
			Required(),
		edge.To("journeys", Journey.Type),
	}
}

// Indexes of the Feature.
func (Feature) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "product_id"),
		index.Fields("org_id", "capability_id"),
		index.Fields("capability_id", "name").Unique(),
	}
}
