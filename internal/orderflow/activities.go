package orderflow

import (
	"context"

	"github.com/google/uuid"

	"github.com/ptflp/gotemporal/internal/orderstatus"
)

const (
	CreateOrderActivityName       = "orderflow.CreateOrder"
	SetPaymentPendingActivityName = "orderflow.SetPaymentPending"
	ConfirmPaymentActivityName    = "orderflow.ConfirmPayment"
	StartShippingActivityName     = "orderflow.StartShipping"
	CompleteOrderActivityName     = "orderflow.CompleteOrder"
	FailOrderActivityName         = "orderflow.FailOrder"
)

type Activities struct {
	CreateOrderFn  func(ctx context.Context, orderID uuid.UUID, amount int, customerID string) error
	UpdateStatusFn func(ctx context.Context, orderID uuid.UUID, status orderstatus.Status, reason *string) error
}

func NewActivities(
	createFn func(ctx context.Context, orderID uuid.UUID, amount int, customerID string) error,
	updateFn func(ctx context.Context, orderID uuid.UUID, status orderstatus.Status, reason *string) error,
) *Activities {
	return &Activities{
		CreateOrderFn:  createFn,
		UpdateStatusFn: updateFn,
	}
}

func (a *Activities) CreateOrder(ctx context.Context, orderID uuid.UUID, amount int, customerID string) error {
	if a.CreateOrderFn == nil {
		return nil
	}
	return a.CreateOrderFn(ctx, orderID, amount, customerID)
}

func (a *Activities) SetPaymentPending(ctx context.Context, orderID uuid.UUID) error {
	return a.updateStatus(ctx, orderID, orderstatus.StatusPaymentPending, nil)
}

func (a *Activities) ConfirmPayment(ctx context.Context, orderID uuid.UUID) error {
	return a.updateStatus(ctx, orderID, orderstatus.StatusPaymentConfirmed, nil)
}

func (a *Activities) StartShipping(ctx context.Context, orderID uuid.UUID) error {
	return a.updateStatus(ctx, orderID, orderstatus.StatusShipping, nil)
}

func (a *Activities) CompleteOrder(ctx context.Context, orderID uuid.UUID) error {
	return a.updateStatus(ctx, orderID, orderstatus.StatusCompleted, nil)
}

func (a *Activities) FailOrder(ctx context.Context, orderID uuid.UUID, reason string) error {
	return a.updateStatus(ctx, orderID, orderstatus.StatusFailed, &reason)
}

func (a *Activities) updateStatus(ctx context.Context, orderID uuid.UUID, status orderstatus.Status, reason *string) error {
	if a.UpdateStatusFn == nil {
		return nil
	}
	return a.UpdateStatusFn(ctx, orderID, status, reason)
}
