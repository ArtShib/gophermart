package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/ArtShib/gophermart.git/internal/config"
	"github.com/ArtShib/gophermart.git/internal/httpclient"
	"github.com/ArtShib/gophermart.git/internal/httpserver"
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
	app.AuthSvc = auth.New(app.Logger, app.Storage, cfg.TokenTTLMIN*time.Minute)
	app.OrderSvc = order.New(app.Logger, app.Storage)
	client := httpclient.New(app.Logger)
	app.AccrualSvc = accrual.New(app.Logger, app.Storage, app.Config.WorkerConfig, client, app.Config.AccrualAddress)
	app.Server = &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: httpserver.New(app.AuthSvc, app.OrderSvc, app.Logger, app.Config),
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
	a.Server.Shutdown(ctx)
	a.AccrualSvc.Stop()
	a.Storage.Close()
}
