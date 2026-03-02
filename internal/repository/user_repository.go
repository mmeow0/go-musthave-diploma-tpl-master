package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/mmeow0/gophermart-bonus/internal/model"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, login, passwordHash string) (*model.User, error) {
	user := &model.User{
		Login:        login,
		PasswordHash: passwordHash,
	}

	query := `
		INSERT INTO users (login, password_hash, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query, login, passwordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrLoginExists
		}
		return nil, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{}

	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	user := &model.User{}

	query := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *PostgresUserRepository) Close() error {
	return nil
}
