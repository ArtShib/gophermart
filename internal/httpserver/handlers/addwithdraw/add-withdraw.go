package addwithdraw

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
	AddWithdraw(ctx context.Context, numOrder int64, userID int64, sum float64) error
}

func New(log *slog.Logger, order Order) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "Order.AddWithdraw"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("received request")

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var requestWithdraw models.RequestWithdraw

		err := json.NewDecoder(r.Body).Decode(&requestWithdraw)
		if err != nil {
			log.Error("failed Unmarshal", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(models.UserIDKey).(int64)
		if !ok || userID == 0 { //0
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := order.AddWithdraw(r.Context(), requestWithdraw.Order, userID, requestWithdraw.Sum); err != nil {
			if errors.Is(err, models.ErrNotValidOrderNumber) {
				log.Error("failed add order", "error", models.ErrNotValidOrderNumber)
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			}
			if errors.Is(err, models.ErrWithdrawBalanceUser) {
				log.Error("there are not enough bonuses to deduct", "error", models.ErrWithdrawBalanceUser)
				http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
				return
			}
			log.Error("failed add order", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
