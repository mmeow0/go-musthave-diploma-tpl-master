package service

import (
	"context"
	"errors"

	"github.com/mmeow0/gophermart-bonus/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrLoginPasswordRequired    = errors.New("login and password are required")
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, login, password string) (int64, error) {
	if login == "" || password == "" {
		return 0, ErrLoginPasswordRequired
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user, err := s.repo.CreateUser(ctx, login, string(passwordHash))
	if err != nil {
		if errors.Is(err, repository.ErrLoginExists) {
			return 0, repository.ErrLoginExists
		}
		return 0, err
	}

	return user.ID, nil
}

func (s *UserService) Login(ctx context.Context, login, password string) (int64, error) {
	if login == "" || password == "" {
		return 0, ErrLoginPasswordRequired
	}

	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return 0, ErrInvalidCredentials
	}

	return user.ID, nil
}
