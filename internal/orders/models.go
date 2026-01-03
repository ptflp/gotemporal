package orders

import (
	"time"

	"github.com/google/uuid"

	"github.com/ptflp/gotemporal/internal/orderstatus"
)

type Order struct {
	ID         uuid.UUID          `json:"id"`
	Status     orderstatus.Status `json:"status"`
	Reason     *string            `json:"reason,omitempty"`
	Amount     int                `json:"amount,omitempty"`
	CustomerID string             `json:"customer_id,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type CreateOrderRequest struct {
	Amount     int    `json:"amount"`
	CustomerID string `json:"customer_id"`
}
