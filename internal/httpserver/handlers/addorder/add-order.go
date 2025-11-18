package addorder

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ArtShib/gophermart.git/internal/models"
	"github.com/go-chi/chi/middleware"
)

type Order interface {
	Add(ctx context.Context, numOrder int64, userID int64) error
}

func New(log *slog.Logger, order Order) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "Order.Add"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("received request")

		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			//log.Error("StatusBadRequest", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		orderNumberStr := string(body)

		orderNumber, err := strconv.ParseInt(orderNumberStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid order number", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(models.UserIDKey).(int64)
		if !ok || userID == 0 { //0
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := order.Add(r.Context(), orderNumber, userID); err != nil {
			if errors.Is(err, models.ErrNotValidOrderNumber) {
				log.Error("failed add order", "error", models.ErrNotValidOrderNumber)
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			}
			if errors.Is(err, models.ErrOrderExists) {
				log.Error("failed add order", "error", models.ErrOrderExists)
				//http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
				w.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, models.ErrOrderExistsOtherUser) {
				log.Error("failed add order", "error", models.ErrOrderExistsOtherUser)
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			}
			log.Error("failed add order", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
