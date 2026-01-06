package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	_ "github.com/lib/pq"
	"github.com/ptflp/gotemporal/ent"
	"github.com/ptflp/gotemporal/internal/config"
	"github.com/ptflp/gotemporal/internal/httpapi"
	"github.com/ptflp/gotemporal/internal/orderflow"
	"github.com/ptflp/gotemporal/internal/orders"
)

func main() {
	app := fx.New(
		fx.WithLogger(func(log *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: log}
		}),
		fx.Provide(
			config.Load,
			newLogger,
			newEntClient,
			newTemporalClient,
			newActivities,
			newWorker,
			newRepository,
			newOrdersService,
			newHTTPServer,
		),
		fx.Invoke(
			startHTTPServer,
			runWorker,
		),
	)

	app.Run()

	if err := app.Err(); err != nil {
		os.Exit(1)
	}
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func newEntClient(lc fx.Lifecycle, cfg config.Config, logger *slog.Logger) (*ent.Client, error) {
	client, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database", "err", err)
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			migrationCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			if err := client.Schema.Create(migrationCtx); err != nil {
				logger.Error("failed to run migrations", "err", err)
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			client.Close()
			return nil
		},
	})

	return client, nil
}

func newTemporalClient(lc fx.Lifecycle, cfg config.Config, logger *slog.Logger) (client.Client, error) {
	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHostPort,
		Namespace: cfg.TemporalNamespace,
	})
	if err != nil {
		logger.Error("failed to connect to temporal", "err", err)
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			temporalClient.Close()
			return nil
		},
	})

	return temporalClient, nil
}

func newActivities(repo *orders.Repository) *orderflow.Activities {
	return orderflow.NewActivities(repo.CreateOrder, repo.UpdateStatus)
}

func newRepository(client *ent.Client) *orders.Repository {
	return orders.NewRepository(client)
}

func newWorker(cfg config.Config, temporalClient client.Client, activities *orderflow.Activities) worker.Worker {
	w := worker.New(temporalClient, cfg.TaskQueue, worker.Options{})
	w.RegisterWorkflow(orderflow.OrderWorkflow)
	w.RegisterActivityWithOptions(activities.CreateOrder, activity.RegisterOptions{Name: orderflow.CreateOrderActivityName})
	w.RegisterActivityWithOptions(activities.SetPaymentPending, activity.RegisterOptions{Name: orderflow.SetPaymentPendingActivityName})
	w.RegisterActivityWithOptions(activities.ConfirmPayment, activity.RegisterOptions{Name: orderflow.ConfirmPaymentActivityName})
	w.RegisterActivityWithOptions(activities.StartShipping, activity.RegisterOptions{Name: orderflow.StartShippingActivityName})
	w.RegisterActivityWithOptions(activities.CompleteOrder, activity.RegisterOptions{Name: orderflow.CompleteOrderActivityName})
	w.RegisterActivityWithOptions(activities.FailOrder, activity.RegisterOptions{Name: orderflow.FailOrderActivityName})

	return w
}

func newOrdersService(repo *orders.Repository, temporalClient client.Client, cfg config.Config) *orders.Service {
	return orders.NewService(repo, temporalClient, cfg.TaskQueue, cfg.PaymentDelay, cfg.ShippingDelay)
}

func newHTTPServer(cfg config.Config, svc *orders.Service) *httpapi.Server {
	return httpapi.NewServer(cfg.HTTPAddr, svc)
}

type workerRunnerParams struct {
	fx.In

	Worker     worker.Worker
	Shutdowner fx.Shutdowner
	Logger     *slog.Logger
}

func runWorker(lc fx.Lifecycle, params workerRunnerParams) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := params.Worker.Run(worker.InterruptCh()); err != nil {
					params.Logger.Error("worker exited with error", "err", err)
					_ = params.Shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Worker.Stop()
			return nil
		},
	})
}

type serverParams struct {
	fx.In

	Server     *httpapi.Server
	Shutdowner fx.Shutdowner
	Logger     *slog.Logger
}

func startHTTPServer(lc fx.Lifecycle, params serverParams) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := params.Server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					params.Logger.Error("http server exited with error", "err", err)
					_ = params.Shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return params.Server.Shutdown(shutdownCtx)
		},
	})
}
