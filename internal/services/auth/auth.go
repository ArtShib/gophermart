package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ArtShib/gophermart.git/internal/lib/jwt"
	"github.com/ArtShib/gophermart.git/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type StoreUser interface {
	SaveUser(ctx context.Context, login string, passHash []byte) (*models.User, error)
	User(ctx context.Context, login string) (*models.User, error)
}

type Auth struct {
	log      *slog.Logger
	store    StoreUser
	tokenTTL time.Duration
}

func New(log *slog.Logger, store StoreUser, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:      log,
		store:    store,
		tokenTTL: tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, login string, password string, secretKey []byte) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("login", login))

	log.Info("login user")

	user, err := a.store.User(ctx, login)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			a.log.Warn("user not found", err)

			return "", fmt.Errorf("%s: %w", op, models.ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", "error", err)

		return "", fmt.Errorf("%s: %w", op, models.ErrInvalidCredentials)
	}

	log.Info("login success")

	token, err := jwt.NewToken(user, a.tokenTTL, secretKey)
	if err != nil {
		a.log.Error("failed to create token", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, login string, pass string, secretKey []byte) (string, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("login", login))

	log.Info("register user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate password hash", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.store.SaveUser(ctx, login, passHash)
	if err != nil {
		a.log.Error("failed to save user", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(user, a.tokenTTL, secretKey)
	if err != nil {
		a.log.Error("failed to create token", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("login success")

	return token, nil
}

func (a *Auth) ParseToken(tokenString string, secretKey []byte) (int64, error) {
	const op = "Auth.ParseToken"

	log := a.log.With(
		slog.String("op", op),
		slog.String("token", tokenString))

	log.Info("parse token")
	token, err := jwt.ParseToken(tokenString, secretKey)
	if err != nil {
		a.log.Error("failed to parse token", "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("login success")
	return token.UserID, nil
}
