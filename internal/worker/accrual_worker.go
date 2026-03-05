package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mmeow0/gophermart-bonus/internal/client"
	"github.com/mmeow0/gophermart-bonus/internal/model"
	"github.com/mmeow0/gophermart-bonus/internal/service"
	"go.uber.org/zap"
)

type AccrualWorker struct {
	orderService    *service.OrderService
	accrualClient   *client.AccrualClient
	logger          *zap.Logger
	stopChan        chan struct{}
	wg              sync.WaitGroup
	interval        time.Duration
	intervalMu      sync.Mutex
	intervalChanged chan struct{}
}

func NewAccrualWorker(
	orderService *service.OrderService,
	accrualClient *client.AccrualClient,
	logger *zap.Logger,
) *AccrualWorker {
	return &AccrualWorker{
		orderService:    orderService,
		accrualClient:   accrualClient,
		logger:          logger,
		stopChan:        make(chan struct{}),
		intervalChanged: make(chan struct{}, 1),
		interval:        5 * time.Second,
	}
}

func (w *AccrualWorker) getInterval() time.Duration {
	w.intervalMu.Lock()
	defer w.intervalMu.Unlock()
	return w.interval
}

func (w *AccrualWorker) setInterval(d time.Duration) {
	w.intervalMu.Lock()
	defer w.intervalMu.Unlock()
	w.interval = d

	// Уведомляем об изменении интервала
	select {
	case w.intervalChanged <- struct{}{}:
	default:
	}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
}

func (w *AccrualWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

func (w *AccrualWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.getInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Accrual worker stopped by context")
			return
		case <-w.stopChan:
			w.logger.Info("Accrual worker stopped")
			return
		case <-w.intervalChanged:
			// Пересоздаём ticker с новым интервалом
			ticker.Stop()
			newInterval := w.getInterval()
			ticker = time.NewTicker(newInterval)
			w.logger.Info("Accrual worker interval updated",
				zap.Duration("new_interval", newInterval),
			)
		case <-ticker.C:
			if err := w.processOrders(ctx); err != nil {
				w.logger.Error("Failed to process orders", zap.Error(err))
			}
		}
	}
}

func (w *AccrualWorker) processOrders(ctx context.Context) error {
	orders, err := w.orderService.GetOrdersForProcessing(ctx, 10)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-w.stopChan:
			return nil
		default:
			if err := w.processOrder(ctx, order); err != nil {
				w.logger.Error("Failed to process order",
					zap.String("order_number", order.Number),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

func (w *AccrualWorker) processOrder(ctx context.Context, order model.Order) error {
	accrualResp, retryAfter, err := w.accrualClient.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		if errors.Is(err, client.ErrTooManyRequests) {
			w.logger.Warn("Too many requests to accrual system",
				zap.Int("retry_after_seconds", retryAfter),
			)
			w.setInterval(time.Duration(retryAfter) * time.Second)
			return nil
		}

		if errors.Is(err, client.ErrNotRegistered) {
			w.logger.Debug("Order not yet registered in accrual system",
				zap.String("order_number", order.Number),
			)
			return nil
		}

		return err
	}

	status := model.OrderStatus(accrualResp.Status)

	switch status {
	case model.OrderStatusRegistered:
		if order.Status == model.OrderStatusNew {
			if err := w.orderService.UpdateOrderStatus(ctx, order.Number, model.OrderStatusProcessing, 0); err != nil {
				return err
			}
		}

	case model.OrderStatusProcessing:
		if order.Status != model.OrderStatusProcessing {
			if err := w.orderService.UpdateOrderStatus(ctx, order.Number, model.OrderStatusProcessing, 0); err != nil {
				return err
			}
		}

	case model.OrderStatusInvalid:
		if err := w.orderService.UpdateOrderStatus(ctx, order.Number, model.OrderStatusInvalid, 0); err != nil {
			return err
		}
		w.logger.Info("Order marked as invalid",
			zap.String("order_number", order.Number),
		)

	case model.OrderStatusProcessed:
		if err := w.orderService.UpdateOrderStatus(ctx, order.Number, model.OrderStatusProcessed, accrualResp.Accrual); err != nil {
			return err
		}
		w.logger.Info("Order processed successfully",
			zap.String("order_number", order.Number),
			zap.Float64("accrual", accrualResp.Accrual),
		)
	}

	return nil
}
