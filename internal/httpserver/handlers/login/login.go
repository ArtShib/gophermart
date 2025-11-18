package login

import (
	"context"
	"encoding/json"
	"errors"

	"log/slog"
	"net/http"

	"github.com/ArtShib/gophermart.git/internal/config"
	"github.com/ArtShib/gophermart.git/internal/models"
	"github.com/go-chi/chi/middleware"
)

type AuthLogin interface {
	Login(ctx context.Context, login string, password string, secretKey []byte) (string, error)
}

func New(log *slog.Logger, authLogin AuthLogin, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "Auth.Login"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("received request")

		var requestUser models.RequestUser

		err := json.NewDecoder(r.Body).Decode(&requestUser)
		if err != nil {
			log.Error("failed Unmarshal", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if requestUser.Login == "" || requestUser.Password == "" {
			log.Error("failed Unmarshal", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		token, err := authLogin.Login(r.Context(), requestUser.Login, requestUser.Password, cfg.SecretKey)
		if err != nil {
			if errors.Is(err, models.ErrInvalidCredentials) {
				log.Error("Invalid Credentials", "error", err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			log.Error("failed login", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
	}
}
