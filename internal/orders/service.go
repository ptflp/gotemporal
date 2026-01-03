package orders

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/ptflp/gotemporal/internal/orderflow"
)

type Service struct {
	repo          *Repository
	temporal      client.Client
	taskQueue     string
	paymentDelay  time.Duration
	shippingDelay time.Duration
}

func NewService(repo *Repository, temporal client.Client, taskQueue string, paymentDelay, shippingDelay time.Duration) *Service {
	return &Service{
		repo:          repo,
		temporal:      temporal,
		taskQueue:     taskQueue,
		paymentDelay:  paymentDelay,
		shippingDelay: shippingDelay,
	}
}

func (s *Service) Create(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	order, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	input := orderflow.OrderWorkflowInput{
		OrderID:       order.ID,
		PaymentDelay:  s.paymentDelay,
		ShippingDelay: s.shippingDelay,
	}

	_, err = s.temporal.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID(order.ID),
		TaskQueue: s.taskQueue,
	}, orderflow.OrderWorkflow, input)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Order, error) {
	return s.repo.Get(ctx, id)
}

func workflowID(id uuid.UUID) string {
	return fmt.Sprintf("order-%s", id.String())
}
