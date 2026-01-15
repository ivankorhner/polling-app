package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Vote holds the schema definition for the Vote entity.
type Vote struct {
	ent.Schema
}

// Fields of the Vote.
func (Vote) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.Int("poll_id").
			Immutable(),
		field.Int("option_id").
			Immutable(),
		field.Int("user_id").
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Vote.
func (Vote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("poll", Poll.Type).
			Ref("votes").
			Field("poll_id").
			Required().
			Unique().
			Immutable(),
		edge.From("option", PollOption.Type).
			Ref("votes").
			Field("option_id").
			Required().
			Unique().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("user", User.Type).
			Ref("votes").
			Field("user_id").
			Required().
			Unique().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Vote - prevents duplicate votes from same user on same poll
func (Vote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "poll_id").
			Unique(),
	}
}
