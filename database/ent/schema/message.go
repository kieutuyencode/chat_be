package schema

import (
	"backend/database/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Message struct {
	ent.Schema
}

func (Message) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "message"},
	}
}

func (Message) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.String("content").Optional(),
		field.Bool("isSeen").Default(false),
		field.Int("conversationId").StorageKey("conversation_id"),
		field.Int("userId").StorageKey("user_id"),
	}
}

func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("conversation", Conversation.Type).
			Ref("messages").Field("conversationId").
			Unique().Required(),
		edge.From("user", User.Type).
			Ref("messages").Field("userId").
			Unique().Required(),
		edge.To("media", MessageMedia.Type),
	}
}
