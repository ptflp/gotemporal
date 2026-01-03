package orderflow

import (
	"context"

	"github.com/google/uuid"

	"github.com/ptflp/gotemporal/internal/orderstatus"
)

const UpdateStatusActivityName = "orderflow.UpdateStatus"

type Activities struct {
	UpdateStatusFn func(ctx context.Context, orderID uuid.UUID, status orderstatus.Status, reason *string) error
}

func NewActivities(fn func(ctx context.Context, orderID uuid.UUID, status orderstatus.Status, reason *string) error) *Activities {
	return &Activities{UpdateStatusFn: fn}
}

func (a *Activities) UpdateStatus(ctx context.Context, orderID uuid.UUID, status orderstatus.Status, reason *string) error {
	if a.UpdateStatusFn == nil {
		return nil
	}
	return a.UpdateStatusFn(ctx, orderID, status, reason)
}
