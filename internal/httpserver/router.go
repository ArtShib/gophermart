package httpserver

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ArtShib/gophermart.git/internal/config"
	"github.com/ArtShib/gophermart.git/internal/httpserver/handlers/addorder"
	"github.com/ArtShib/gophermart.git/internal/httpserver/handlers/addwithdraw"
	"github.com/ArtShib/gophermart.git/internal/httpserver/handlers/getbalance"
	"github.com/ArtShib/gophermart.git/internal/httpserver/handlers/getorder"
	"github.com/ArtShib/gophermart.git/internal/httpserver/handlers/getwithdrawals"
	"github.com/ArtShib/gophermart.git/internal/httpserver/handlers/login"
	"github.com/ArtShib/gophermart.git/internal/httpserver/handlers/register"
	mwAuth "github.com/ArtShib/gophermart.git/internal/httpserver/middleware/auth"
	mwLogger "github.com/ArtShib/gophermart.git/internal/httpserver/middleware/logger"
	"github.com/ArtShib/gophermart.git/internal/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type AuthService interface {
	RegisterNewUser(ctx context.Context, login string, pass string, secretKey []byte) (string, error)
	Login(ctx context.Context, login string, password string, secretKey []byte) (string, error)
	ParseToken(tokenString string, secretKey []byte) (int64, error)
}

type Order interface {
	Add(ctx context.Context, numOrder int64, userID int64) error
	Get(ctx context.Context, userID int64) (models.OrderArray, error)
	Balance(ctx context.Context, userID int64) (*models.Balance, error)
	Withdrawals(ctx context.Context, userID int64) (models.WithdrawalsArray, error)
	AddWithdraw(ctx context.Context, numOrder int64, userID int64, sum float64) error
}

func New(svc AuthService, order Order, log *slog.Logger, cfg *config.Config) http.Handler {

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(mwLogger.New(log))

	mux.Route("/api/user", func(r chi.Router) {
		//r.Use(gzipHandle)
		r.Post("/register", register.New(log, svc, cfg))
		r.Post("/login", login.New(log, svc, cfg))
	})

	mux.Group(func(r chi.Router) {
		//r.Use(gzipHandle)
		r.Use(mwAuth.New(log, svc, cfg))
		r.Post("/api/user/orders", addorder.New(log, order))
		r.Get("/api/user/orders", getorder.New(log, order))
		r.Get("/api/user/balance", getbalance.New(log, order))
		r.Get("/api/user/withdrawals", getwithdrawals.New(log, order))
		r.Post("/api/user/balance/withdraw", addwithdraw.New(log, order))
	})
	return mux
}
