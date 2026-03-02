package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/mmeow0/gophermart-bonus/internal/model"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
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

	err := r.db.QueryRowContext(ctx, query, number, userID, model.OrderStatusNew).Scan(
		&order.ID,
		&order.UploadedAt,
	)
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

	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
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

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`

	result, err := r.db.ExecContext(ctx, query, status, accrual, number)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresOrderRepository) GetOrdersForProcessing(ctx context.Context, limit int) ([]model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at 
		FROM orders 
		WHERE status IN ($1, $2)
		ORDER BY uploaded_at ASC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, model.OrderStatusNew, model.OrderStatusProcessing, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresOrderRepository) Close() error {
	return nil
}
