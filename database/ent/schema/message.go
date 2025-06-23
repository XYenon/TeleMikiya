package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/xyenon/telemikiya/types"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int("msg_id"),
		field.Int64("dialog_id"),
		field.String("text").
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Other("text_embedding", pgvector.Vector{}).
			SchemaType(map[string]string{dialect.Postgres: "vector(%d)"}).
			Optional(),
		field.Bool("has_media"),
		field.JSON("media_info", &types.MediaInfo{}),
		field.Time("sent_at"),
	}
}

// Indexes of the Message.
func (Message) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("msg_id", "dialog_id").Unique(),
		index.Fields("text").Annotations(entsql.IndexType("pgroonga")),
		index.Fields("text_embedding").
			Annotations(
				entsql.IndexType("vchordrq"),
				entsql.OpClass("vector_cosine_ops"),
			),
		index.Fields("sent_at"),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dialog", Dialog.Type).Ref("messages").Field("dialog_id").Unique().Required(),
	}
}
