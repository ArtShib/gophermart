package models

import "errors"

var (
	ErrUserExists           = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrOrderExists          = errors.New("order already exists")
	ErrOrderExistsOtherUser = errors.New("order already exists other user")
	ErrNotValidOrderNumber  = errors.New("order number is not valid")
	ErrOrderEmpty           = errors.New("order is empty")
	ErrWithdrawalsEmpty     = errors.New("withdrawals is empty")
	ErrWithdrawBalanceUser  = errors.New("there are not enough bonuses to deduct")
	ErrOrdersInWorkIsEmpty  = errors.New("list of orders is empty")
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)
