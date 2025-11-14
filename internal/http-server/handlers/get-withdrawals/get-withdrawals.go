package get_withdrawals

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ArtShib/gophermart.git/internal/models"
	"github.com/go-chi/chi/middleware"
)

type Order interface {
	Withdrawals(ctx context.Context, userID int64) (models.WithdrawalsArray, error)
}

func New(log *slog.Logger, order Order) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "Withdrawals.Get"

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

		withdrawals, err := order.Withdrawals(r.Context(), userID)
		if err != nil {
			if errors.Is(err, models.ErrWithdrawalsEmpty) {
				log.Error("Withdrawals is empty", "error", models.ErrWithdrawalsEmpty)
				http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
				return
			}
			log.Error("get balance", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(withdrawals); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
