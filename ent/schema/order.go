package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Order holds the schema definition for the Order entity.
type Order struct {
	ent.Schema
}

// Fields of the Order.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Enum("status").
			Values(
				"created",
				"payment_pending",
				"payment_confirmed",
				"shipping",
				"completed",
				"failed",
				"cancelled",
			),
		field.String("reason").
			Optional().
			Nillable(),
		field.Int("amount").
			Optional(),
		field.String("customer_id").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Order.
func (Order) Edges() []ent.Edge {
	return nil
}

