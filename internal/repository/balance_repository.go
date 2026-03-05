package repository

import (
	"context"
	"database/sql"

	"github.com/mmeow0/gophermart-bonus/internal/model"
)

type PostgresBalanceRepository struct {
	*BaseRepository[model.Withdrawal]
}

func NewPostgresBalanceRepository(db *sql.DB) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{
		BaseRepository: NewBaseRepository[model.Withdrawal](db),
	}
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
	err = r.QueryRow(ctx, query, func(row *sql.Row) error {
		return row.Scan(&accrued, &withdrawn)
	}, model.OrderStatusProcessed, userID)

	if err != nil {
		return 0, 0, err
	}

	current = accrued - withdrawn
	return current, withdrawn, nil
}

func (r *PostgresBalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	current, _, err := r.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if current < sum {
		return ErrInsufficientFunds
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

	withdrawals := []model.Withdrawal{}
	err := r.Query(ctx, query, func(rows *sql.Rows) error {
		var w model.Withdrawal
		if err := rows.Scan(&w.ID, &w.UserID, &w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return err
		}
		withdrawals = append(withdrawals, w)
		return nil
	}, userID)

	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (r *PostgresBalanceRepository) Close() error {
	return r.BaseRepository.Close()
}
