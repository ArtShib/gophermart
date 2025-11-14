package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ArtShib/gophermart.git/internal/config"
	"github.com/ArtShib/gophermart.git/internal/models"
)

type ParseAuth interface {
	ParseToken(tokenString string, secretKey []byte) (int64, error)
}

func New(log *slog.Logger, auth ParseAuth, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("auth middleware enabled")
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			userID, err := auth.ParseToken(authHeader, cfg.SecretKey)

			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), models.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
