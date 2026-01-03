package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/ptflp/gotemporal/internal/orders"
)

type Server struct {
	httpServer *http.Server
	handler    *handler
}

func NewServer(addr string, svc *orders.Service) *Server {
	h := &handler{svc: svc}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Post("/orders", h.createOrder)
	r.Get("/orders/{id}", h.getOrder)
	registerDocsRoutes(r)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		handler: h,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

type handler struct {
	svc *orders.Service
}

type createOrderRequest struct {
	Amount     int    `json:"amount"`
	CustomerID string `json:"customer_id"`
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

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	order, err := h.svc.Create(r.Context(), orders.CreateOrderRequest{
		Amount:     req.Amount,
		CustomerID: req.CustomerID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toResponse(order))
}

func (h *handler) getOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	order, err := h.svc.Get(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, orders.ErrOrderNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toResponse(order))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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
