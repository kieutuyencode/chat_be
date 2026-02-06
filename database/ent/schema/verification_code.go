package schema

import (
	"backend/database/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type VerificationCode struct {
	ent.Schema
}

func (VerificationCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "verification_code"},
	}
}

func (VerificationCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code"),
	}
}

func (VerificationCode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (VerificationCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("code"),
		field.Time("expiresAt").StorageKey("expires_at"),
		field.Int("userId").StorageKey("user_id"),
	}
}

func (VerificationCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("verificationCodes").Field("userId").Unique().Required(),
	}
}
