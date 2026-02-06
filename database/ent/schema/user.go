package schema

import (
	"backend/database/ent/schema/mixin"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user"},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("fullname").MaxLen(50),
		field.String("email").MaxLen(100),
		field.String("phone").MaxLen(50).Optional(),
		field.String("avatar").Optional(),
		field.Bool("isActive").StorageKey("is_active").Default(false),
		field.Time("lastActiveAt").StorageKey("last_active_at").Default(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("verificationCodes", VerificationCode.Type),
		edge.To("conversationMembers", ConversationMember.Type),
		edge.To("messages", Message.Type),
	}
}
