package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ArtShib/gophermart.git/internal/config"
	http_client "github.com/ArtShib/gophermart.git/internal/http-client"
	http_server "github.com/ArtShib/gophermart.git/internal/http-server"
	liblog "github.com/ArtShib/gophermart.git/internal/lib/logger"
	"github.com/ArtShib/gophermart.git/internal/services/accrual"
	"github.com/ArtShib/gophermart.git/internal/services/auth"
	"github.com/ArtShib/gophermart.git/internal/services/order"
	"github.com/ArtShib/gophermart.git/internal/storage"
)

type App struct {
	Logger     *slog.Logger
	Storage    storage.Storage
	Server     *http.Server
	Config     *config.Config
	AuthSvc    *auth.Auth
	OrderSvc   *order.Order
	AccrualSvc *accrual.ClientAccrual
}

func NewApp(cfg *config.Config, store *storage.Storage) *App {
	app := &App{
		Config:  cfg,
		Storage: *store,
	}
	app.Logger = liblog.New()
	app.AuthSvc = auth.New(app.Logger, app.Storage, 10000000)
	app.OrderSvc = order.New(app.Logger, app.Storage)
	client := http_client.New(app.Logger)
	app.AccrualSvc = accrual.New(app.Logger, app.Storage, cfg.WorkerConfig, client, cfg.AccrualAddress)
	app.Server = &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: http_server.New(app.AuthSvc, app.OrderSvc, app.Logger, cfg),
	}
	return app
}

func (a *App) Run(ctx context.Context) {
	a.AccrualSvc.Start(ctx)
	go func() {
		if err := a.Server.ListenAndServe(); err != nil {
			a.Logger.Error(err.Error())
		}

	}()
}

func (a *App) Stop(ctx context.Context) {

	a.AccrualSvc.Stop()
	a.Storage.Close()
	a.Server.Shutdown(ctx)
}
