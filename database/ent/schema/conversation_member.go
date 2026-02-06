package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ConversationMember struct {
	ent.Schema
}

func (ConversationMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "conversation_member"},
	}
}

func (ConversationMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("conversationId", "userId").Unique(),
	}
}

func (ConversationMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int("conversationId").StorageKey("conversation_id"),
		field.Int("userId").StorageKey("user_id"),
	}
}

func (ConversationMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("conversation", Conversation.Type).
			Ref("members").Field("conversationId").
			Unique().Required(),
		edge.From("user", User.Type).
			Ref("conversationMembers").Field("userId").
			Unique().Required(),
	}
}
