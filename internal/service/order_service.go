package service

import (
	"context"
	"errors"

	"github.com/mmeow0/gophermart-bonus/internal/model"
	"github.com/mmeow0/gophermart-bonus/internal/repository"
)

type OrderService struct {
	orderRepo   repository.OrderRepository
	balanceRepo repository.BalanceRepository
}

func NewOrderService(orderRepo repository.OrderRepository, balanceRepo repository.BalanceRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, number string, userID int64) error {
	_, err := s.orderRepo.CreateOrder(ctx, number, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return repository.ErrConflict
		}
		if errors.Is(err, repository.ErrOrderExists) {
			return repository.ErrOrderExists
		}
		return err
	}
	return nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.orderRepo.GetUserOrders(ctx, userID)
}

func (s *OrderService) GetBalance(ctx context.Context, userID int64) (current float64, withdrawn float64, err error) {
	return s.balanceRepo.GetBalance(ctx, userID)
}

func (s *OrderService) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	return s.balanceRepo.CreateWithdrawal(ctx, userID, orderNumber, sum)
}

func (s *OrderService) GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	return s.balanceRepo.GetUserWithdrawals(ctx, userID)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error {
	return s.orderRepo.UpdateOrderStatus(ctx, number, status, accrual)
}

func (s *OrderService) GetOrdersForProcessing(ctx context.Context, limit int) ([]model.Order, error) {
	return s.orderRepo.GetOrdersForProcessing(ctx, limit)
}
