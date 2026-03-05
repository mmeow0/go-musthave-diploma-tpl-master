package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/mmeow0/gophermart-bonus/internal/model"
)

type PostgresOrderRepository struct {
	*BaseRepository[model.Order]
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		BaseRepository: NewBaseRepository[model.Order](db),
	}
}

func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, number string, userID int64) (*model.Order, error) {
	order := &model.Order{
		Number: number,
		UserID: userID,
		Status: model.OrderStatusNew,
	}

	query := `
		INSERT INTO orders (number, user_id, status, uploaded_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, uploaded_at
	`

	err := r.QueryRow(ctx, query, func(row *sql.Row) error {
		return row.Scan(&order.ID, &order.UploadedAt)
	}, number, userID, model.OrderStatusNew)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			existingOrder, getErr := r.GetOrderByNumber(ctx, number)
			if getErr != nil {
				return nil, getErr
			}
			if existingOrder.UserID != userID {
				return nil, ErrOrderExists
			}
			return existingOrder, ErrConflict
		}
		return nil, err
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	order := &model.Order{}

	query := `SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1`

	err := r.QueryRow(ctx, query, func(row *sql.Row) error {
		return row.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	}, number)

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY uploaded_at DESC
	`

	orders := []model.Order{}
	err := r.Query(ctx, query, func(rows *sql.Rows) error {
		var order model.Order
		if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return err
		}
		orders = append(orders, order)
		return nil
	}, userID)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`
	return r.ExecWithRowsAffected(ctx, query, status, accrual, number)
}

func (r *PostgresOrderRepository) GetOrdersForProcessing(ctx context.Context, limit int) ([]model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at 
		FROM orders 
		WHERE status IN ($1, $2)
		ORDER BY uploaded_at ASC
		LIMIT $3
	`

	orders := []model.Order{}
	err := r.Query(ctx, query, func(rows *sql.Rows) error {
		var order model.Order
		if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return err
		}
		orders = append(orders, order)
		return nil
	}, model.OrderStatusNew, model.OrderStatusProcessing, limit)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresOrderRepository) Close() error {
	return r.BaseRepository.Close()
}
