package schema

import (
	"backend/database/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
)

type Conversation struct {
	ent.Schema
}

func (Conversation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "conversation"},
	}
}

func (Conversation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (Conversation) Fields() []ent.Field {
	return []ent.Field{}
}

func (Conversation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", ConversationMember.Type),
		edge.To("messages", Message.Type),
	}
}
