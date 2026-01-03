// Package orderflow contains the Temporal workflow for order processing.
package orderflow

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type OrderWorkflowInput struct {
	OrderID       uuid.UUID
	Amount        int
	CustomerID    string
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

	failOrder := func(reason string) error {
		return workflow.ExecuteActivity(ctx, FailOrderActivityName, input.OrderID, reason).Get(ctx, nil)
	}

	if err := workflow.ExecuteActivity(ctx, CreateOrderActivityName, input.OrderID, input.Amount, input.CustomerID).Get(ctx, nil); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, SetPaymentPendingActivityName, input.OrderID, input.PaymentDelay).Get(ctx, nil); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, ConfirmPaymentActivityName, input.OrderID, input.PaymentDelay).Get(ctx, nil); err != nil {
		_ = failOrder(err.Error())
		return err
	}

	if err := workflow.ExecuteActivity(ctx, StartShippingActivityName, input.OrderID, input.ShippingDelay).Get(ctx, nil); err != nil {
		_ = failOrder(err.Error())
		return err
	}

	if err := workflow.ExecuteActivity(ctx, CompleteOrderActivityName, input.OrderID, input.ShippingDelay).Get(ctx, nil); err != nil {
		_ = failOrder(err.Error())
		return err
	}

	return nil
}
