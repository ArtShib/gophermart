package order

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ArtShib/gophermart.git/internal/lib/luhn"
	"github.com/ArtShib/gophermart.git/internal/models"
)

type StoreOrder interface {
	AddOrder(ctx context.Context, numOrder int64, uploaded int64, userID int64) error
	GetOrder(ctx context.Context, userID int64) (models.OrderArray, error)
	GetBalance(ctx context.Context, userID int64) (*models.Balance, error)
	GetWithdrawals(ctx context.Context, userID int64) (models.WithdrawalsArray, error)
	AddWithdraw(ctx context.Context, numOrder int64, userID int64, sum float64, processed int64) error
	GetOrdersInWork(ctx context.Context) (models.OrderArray, error)
}

type Order struct {
	log   *slog.Logger
	store StoreOrder
}

func New(log *slog.Logger, store StoreOrder) *Order {
	return &Order{
		log:   log,
		store: store,
	}
}

func (o *Order) Add(ctx context.Context, numOrder int64, userID int64) error {
	const op = "Order.AddOrder"

	currentTime := time.Now().Unix()

	log := o.log.With(
		slog.String("op", op),
		slog.String("number", fmt.Sprintf("%v", numOrder)),
		slog.String("user_id", fmt.Sprintf("%v", userID)))

	log.Info("add order")

	if !luhn.Valid(numOrder) {
		o.log.Error("order number is not valid", "error", models.ErrNotValidOrderNumber)
		return models.ErrNotValidOrderNumber
	}
	err := o.store.AddOrder(ctx, numOrder, currentTime, userID)
	return err
}

func (o *Order) Get(ctx context.Context, userID int64) (models.OrderArray, error) {
	const op = "Order.GetOrder"

	log := o.log.With(
		slog.String("op", op),
		slog.String("user_id", fmt.Sprintf("%v", userID)))

	log.Info("get order")

	return o.store.GetOrder(ctx, userID)
}

func (o *Order) GetOrdersInWork(ctx context.Context) (models.OrderArray, error) {
	const op = "Order.GetOrdersInWork"

	log := o.log.With(
		slog.String("op", op))

	log.Info("get Orders In Work")

	return o.store.GetOrdersInWork(ctx)
}

func (o *Order) Balance(ctx context.Context, userID int64) (*models.Balance, error) {
	const op = "Order.GetBalance"

	log := o.log.With(
		slog.String("op", op),
		slog.String("user_id", fmt.Sprintf("%v", userID)))

	log.Info("get balance")

	return o.store.GetBalance(ctx, userID)
}

func (o *Order) Withdrawals(ctx context.Context, userID int64) (models.WithdrawalsArray, error) {
	const op = "Order.GetWithdrawals"

	log := o.log.With(
		slog.String("op", op),
		slog.String("user_id", fmt.Sprintf("%v", userID)))

	log.Info("get withdrawals")

	return o.store.GetWithdrawals(ctx, userID)
}

func (o *Order) AddWithdraw(ctx context.Context, numOrder int64, userID int64, sum float64) error {
	const op = "Order.AddWithdrawal"

	currentTime := time.Now().Unix()

	log := o.log.With(
		slog.String("op", op),
		slog.String("number", fmt.Sprintf("%v", numOrder)),
		slog.String("user_id", fmt.Sprintf("%v", userID)))

	log.Info("add withdrawal")

	if !luhn.Valid(numOrder) {
		o.log.Error("order number is not valid", "error", models.ErrNotValidOrderNumber)
		return models.ErrNotValidOrderNumber
	}
	err := o.store.AddWithdraw(ctx, numOrder, userID, sum, currentTime)
	return err
}
