package getbalance

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ArtShib/gophermart.git/internal/models"
	"github.com/go-chi/chi/middleware"
)

type Order interface {
	Balance(ctx context.Context, userID int64) (*models.Balance, error)
}

func New(log *slog.Logger, order Order) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "Balance.Get"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("received request")

		userID, ok := r.Context().Value(models.UserIDKey).(int64)
		if !ok || userID == 0 { //0
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		getBalance, err := order.Balance(r.Context(), userID)
		if err != nil {
			log.Error("get balance", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(getBalance); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
