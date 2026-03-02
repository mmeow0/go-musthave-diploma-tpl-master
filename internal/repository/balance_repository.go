package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mmeow0/gophermart-bonus/internal/model"
)

type PostgresBalanceRepository struct {
	db *sql.DB
}

func NewPostgresBalanceRepository(db *sql.DB) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{db: db}
}

func (r *PostgresBalanceRepository) GetBalance(ctx context.Context, userID int64) (current float64, withdrawn float64, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = $1 THEN accrual ELSE 0 END), 0) as accrued,
			COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $2), 0) as withdrawn
		FROM orders
		WHERE user_id = $2
	`

	var accrued float64
	err = r.db.QueryRowContext(ctx, query, model.OrderStatusProcessed, userID).Scan(&accrued, &withdrawn)
	if err != nil {
		return 0, 0, err
	}

	current = accrued - withdrawn
	return current, withdrawn, nil
}

func (r *PostgresBalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	current, _, err := r.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if current < sum {
		return errors.New("insufficient funds")
	}

	query := `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err = tx.ExecContext(ctx, query, userID, orderNumber, sum)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresBalanceRepository) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	query := `
		SELECT id, user_id, order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := []model.Withdrawal{}
	for rows.Next() {
		var w model.Withdrawal
		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.OrderNumber,
			&w.Sum,
			&w.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (r *PostgresBalanceRepository) Close() error {
	return nil
}
