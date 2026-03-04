package repository

import (
	"context"
	"errors"

	"github.com/mmeow0/gophermart-bonus/internal/model"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrLoginExists       = errors.New("login already exists")
	ErrOrderExists       = errors.New("order already exists")
	ErrInvalidStatus     = errors.New("invalid order status")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (*model.User, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
	Close() error
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, number string, userID int64) (*model.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error
	GetOrdersForProcessing(ctx context.Context, limit int) ([]model.Order, error)
	Close() error
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (current float64, withdrawn float64, err error)
	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
	Close() error
}
