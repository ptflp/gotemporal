package orders

import (
	"context"
	"errors"
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
	orderID := uuid.New()

	input := orderflow.OrderWorkflowInput{
		OrderID:       orderID,
		Amount:        req.Amount,
		CustomerID:    req.CustomerID,
		PaymentDelay:  s.paymentDelay,
		ShippingDelay: s.shippingDelay,
	}

	_, err := s.temporal.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID(orderID),
		TaskQueue: s.taskQueue,
	}, orderflow.OrderWorkflow, input)
	if err != nil {
		return nil, err
	}

	return s.waitForOrder(ctx, orderID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Order, error) {
	return s.repo.Get(ctx, id)
}

func workflowID(id uuid.UUID) string {
	return fmt.Sprintf("order-%s", id.String())
}

func (s *Service) waitForOrder(ctx context.Context, id uuid.UUID) (*Order, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("order not created by workflow: %w", timeoutCtx.Err())
		case <-ticker.C:
			order, err := s.repo.Get(ctx, id)
			if err == nil {
				return order, nil
			}
			if !errors.Is(err, ErrOrderNotFound) {
				return nil, err
			}
		}
	}
}
