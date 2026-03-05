package repository

import (
	"context"
	"database/sql"
	"errors"
)

// BaseRepository предоставляет общую функциональность для работы с БД
type BaseRepository[T any] struct {
	db *sql.DB
}

// NewBaseRepository создаёт новый базовый репозиторий
func NewBaseRepository[T any](db *sql.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// QueryRow выполняет запрос и возвращает одну строку с обработкой ErrNoRows
func (r *BaseRepository[T]) QueryRow(ctx context.Context, query string, scanFn func(*sql.Row) error, args ...interface{}) error {
	row := r.db.QueryRowContext(ctx, query, args...)
	err := scanFn(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Query выполняет запрос и возвращает несколько строк
func (r *BaseRepository[T]) Query(ctx context.Context, query string, scanFn func(*sql.Rows) error, args ...interface{}) error {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := scanFn(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Exec выполняет запрос без возврата данных
func (r *BaseRepository[T]) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.db.ExecContext(ctx, query, args...)
}

// ExecWithRowsAffected выполняет запрос и проверяет количество затронутых строк
func (r *BaseRepository[T]) ExecWithRowsAffected(ctx context.Context, query string, args ...interface{}) error {
	result, err := r.db.ExecContext(ctx, query, args...)
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

// BeginTx начинает транзакцию
func (r *BaseRepository[T]) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

// GetDB возвращает базу данных для специфичных операций
func (r *BaseRepository[T]) GetDB() *sql.DB {
	return r.db
}

// Close закрывает соединение (в нашем случае ничего не делает)
func (r *BaseRepository[T]) Close() error {
	return nil
}
