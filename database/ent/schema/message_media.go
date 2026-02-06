package schema

import (
	"backend/database/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type MessageMedia struct {
	ent.Schema
}

func (MessageMedia) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "message_media"},
	}
}

func (MessageMedia) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (MessageMedia) Fields() []ent.Field {
	return []ent.Field{
		field.String("src"),
		field.Int("messageId").StorageKey("message_id"),
	}
}

func (MessageMedia) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("message", Message.Type).
			Ref("media").Field("messageId").
			Unique().Required(),
	}
}
