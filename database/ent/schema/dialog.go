package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/xyenon/telemikiya/types"
)

// Dialog holds the schema definition for the Dialog entity.
type Dialog struct {
	ent.Schema
}

// Fields of the Dialog.
func (Dialog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("title"),
		field.Enum("type").GoType(types.DialogType("")),
		field.Time("updated_at").Default(time.Now),
	}
}

// Edges of the Dialog.
func (Dialog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("messages", Message.Type),
	}
}
