package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/ptflp/gotemporal/ent"
	"github.com/ptflp/gotemporal/internal/config"
	"github.com/ptflp/gotemporal/internal/httpapi"
	"github.com/ptflp/gotemporal/internal/orderflow"
	"github.com/ptflp/gotemporal/internal/orders"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	entClient, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer entClient.Close()

	if err := entClient.Schema.Create(ctx); err != nil {
		logger.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHostPort,
		Namespace: cfg.TemporalNamespace,
	})
	if err != nil {
		logger.Error("failed to connect to temporal", "err", err)
		os.Exit(1)
	}
	defer temporalClient.Close()

	repo := orders.NewRepository(entClient)
	activities := orderflow.NewActivities(repo.UpdateStatus)

	w := worker.New(temporalClient, cfg.TaskQueue, worker.Options{})
	w.RegisterWorkflow(orderflow.OrderWorkflow)
	w.RegisterActivityWithOptions(activities.UpdateStatus, activity.RegisterOptions{Name: orderflow.UpdateStatusActivityName})

	service := orders.NewService(repo, temporalClient, cfg.TaskQueue, cfg.PaymentDelay, cfg.ShippingDelay)
	server := httpapi.NewServer(cfg.HTTPAddr, service)

	errCh := make(chan error, 2)

	go func() {
		if err := w.Run(worker.InterruptCh()); err != nil {
			errCh <- err
		}
	}()

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	logger.Info("app started", "http", cfg.HTTPAddr, "taskQueue", cfg.TaskQueue)

	select {
	case <-ctx.Done():
		logger.Info("shutting down")
	case err := <-errCh:
		logger.Error("service error", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
	w.Stop()
}
