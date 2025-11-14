package storage

import (
	"context"
	"fmt"

	"github.com/ArtShib/gophermart.git/internal/models"
	"github.com/ArtShib/gophermart.git/internal/storage/postgres"
)

type Storage interface {
	Close() error
	SaveUser(ctx context.Context, login string, passHash []byte) (*models.User, error)
	User(ctx context.Context, login string) (*models.User, error)
	AddOrder(ctx context.Context, numOrder int64, uploaded int64, userID int64) error
	GetOrder(ctx context.Context, userID int64) (models.OrderArray, error)
	GetBalance(ctx context.Context, userID int64) (*models.Balance, error)
	GetWithdrawals(ctx context.Context, userID int64) (models.WithdrawalsArray, error)
	AddWithdraw(ctx context.Context, numOrder int64, userID int64, sum float64, processed int64) error
	GetOrdersInWork(ctx context.Context) (models.OrderArray, error)
	UpdateOrdersBatch(ctx context.Context, orders models.ResAccrualOrderArray) error
}

func New(ctx context.Context, dsn string) (Storage, error) {
	const op = "storage.NewStorage"

	store, err := postgres.NewPostgresStore(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return store, nil
}
