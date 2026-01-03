// Package httpapi provides the HTTP API for orders, backed by Temporal workflows.
package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/ptflp/gotemporal/internal/orders"
)

type Server struct {
	httpServer *http.Server
	api        huma.API
}

func NewServer(addr string, svc *orders.Service) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	cfg := huma.DefaultConfig("Order Service API", "1.0.0")
	cfg.Info.Description = "Minimal REST API to create orders and read their status. Orchestration is handled by Temporal; statuses are persisted in Postgres."
	// Drop default schema link transformer to avoid `$schema` fields in responses/docs.
	cfg.CreateHooks = nil
	api := humachi.New(router, cfg)

	registerRoutes(api, svc)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		api: api,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

type createOrderRequest struct {
	Amount     int    `json:"amount"`
	CustomerID string `json:"customer_id"`
}

type createOrderInput struct {
	Body createOrderRequest `required:"true"`
}

type createOrderOutput struct {
	Body orderResponse
}

type getOrderInput struct {
	ID uuid.UUID `path:"id"`
}

type getOrderOutput struct {
	Body orderResponse
}

type orderResponse struct {
	ID         uuid.UUID `json:"id"`
	Status     string    `json:"status"`
	Reason     *string   `json:"reason,omitempty"`
	Amount     int       `json:"amount,omitempty"`
	CustomerID string    `json:"customer_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func registerRoutes(api huma.API, svc *orders.Service) {
	huma.Post(api, "/orders", func(ctx context.Context, input *createOrderInput) (*createOrderOutput, error) {
		order, err := svc.Create(ctx, orders.CreateOrderRequest{
			Amount:     input.Body.Amount,
			CustomerID: input.Body.CustomerID,
		})
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, err.Error())
		}

		return &createOrderOutput{Body: toResponse(order)}, nil
	}, huma.OperationTags("orders"), func(op *huma.Operation) {
		op.DefaultStatus = http.StatusCreated
	})

	huma.Get(api, "/orders/{id}", func(ctx context.Context, input *getOrderInput) (*getOrderOutput, error) {
		order, err := svc.Get(ctx, input.ID)
		if err != nil {
			if errors.Is(err, orders.ErrOrderNotFound) {
				return nil, huma.NewError(http.StatusNotFound, "order not found")
			}
			return nil, huma.NewError(http.StatusInternalServerError, err.Error())
		}

		return &getOrderOutput{Body: toResponse(order)}, nil
	}, huma.OperationTags("orders"))
}

func toResponse(o *orders.Order) orderResponse {
	return orderResponse{
		ID:         o.ID,
		Status:     string(o.Status),
		Reason:     o.Reason,
		Amount:     o.Amount,
		CustomerID: o.CustomerID,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}
