package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PollOption holds the schema definition for the PollOption entity.
type PollOption struct {
	ent.Schema
}

// Fields of the PollOption.
func (PollOption) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.Int("poll_id").
			Immutable(),
		field.String("text").
			NotEmpty(),
		field.Int("vote_count").
			Default(0).
			NonNegative(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the PollOption.
func (PollOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("poll", Poll.Type).
			Ref("options").
			Field("poll_id").
			Unique().
			Required().
			Immutable(),
		edge.To("votes", Vote.Type),
	}
}
