package orders

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ptflp/gotemporal/ent"
	"github.com/ptflp/gotemporal/ent/order"
	"github.com/ptflp/gotemporal/internal/orderstatus"
)

var ErrOrderNotFound = errors.New("order not found")

type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) Create(ctx context.Context, id uuid.UUID, req CreateOrderRequest) (*Order, error) {
	created, err := r.client.Order.
		Create().
		SetID(id).
		SetAmount(req.Amount).
		SetCustomerID(req.CustomerID).
		SetStatus(order.Status(orderstatus.StatusCreated)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapModel(created), nil
}

func (r *Repository) CreateOrder(ctx context.Context, id uuid.UUID, amount int, customerID string) error {
	_, err := r.Create(ctx, id, CreateOrderRequest{
		Amount:     amount,
		CustomerID: customerID,
	})
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status orderstatus.Status, reason *string) error {
	_, err := r.client.Order.
		UpdateOneID(id).
		SetStatus(order.Status(status)).
		SetNillableReason(reason).
		Save(ctx)
	return err
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*Order, error) {
	found, err := r.client.Order.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return mapModel(found), nil
}

func mapModel(o *ent.Order) *Order {
	return &Order{
		ID:         o.ID,
		Status:     orderstatus.Status(o.Status),
		Reason:     o.Reason,
		Amount:     o.Amount,
		CustomerID: o.CustomerID,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}
