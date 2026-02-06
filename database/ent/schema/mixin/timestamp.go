package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type Timestamp struct {
	mixin.Schema
}

func (Timestamp) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createdAt").
			StorageKey("created_at").
			Default(time.Now),
		field.Time("updatedAt").
			StorageKey("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

type CreatedAt struct {
	mixin.Schema
}

func (CreatedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now),
	}
}

type UpdatedAt struct {
	mixin.Schema
}

func (UpdatedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}
