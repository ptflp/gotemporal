package orderflow

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/ptflp/gotemporal/internal/orderstatus"
)

type OrderWorkflowInput struct {
	OrderID       uuid.UUID
	PaymentDelay  time.Duration
	ShippingDelay time.Duration
}

func OrderWorkflow(ctx workflow.Context, input OrderWorkflowInput) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	update := func(status orderstatus.Status, reason *string) error {
		return workflow.ExecuteActivity(ctx, UpdateStatusActivityName, input.OrderID, status, reason).Get(ctx, nil)
	}

	if err := update(orderstatus.StatusPaymentPending, nil); err != nil {
		return err
	}

	if input.PaymentDelay > 0 {
		if err := workflow.Sleep(ctx, input.PaymentDelay); err != nil {
			return err
		}
	}

	if err := update(orderstatus.StatusPaymentConfirmed, nil); err != nil {
		return err
	}

	if err := update(orderstatus.StatusShipping, nil); err != nil {
		return err
	}

	if input.ShippingDelay > 0 {
		if err := workflow.Sleep(ctx, input.ShippingDelay); err != nil {
			return err
		}
	}

	if err := update(orderstatus.StatusCompleted, nil); err != nil {
		return err
	}

	return nil
}
