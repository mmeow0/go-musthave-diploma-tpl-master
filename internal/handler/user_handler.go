package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mmeow0/gophermart-bonus/internal/middleware"
	"github.com/mmeow0/gophermart-bonus/internal/model"
	"github.com/mmeow0/gophermart-bonus/internal/repository"
	"github.com/mmeow0/gophermart-bonus/internal/service"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
	secretKey   string
	logger      *zap.Logger
}

func NewUserHandler(userService *service.UserService, secretKey string, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		secretKey:   secretKey,
		logger:      logger,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.userService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrLoginExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.logger.Error("Failed to register user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := middleware.GenerateToken(userID, h.secretKey)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.userService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Debug("Login failed", zap.String("login", req.Login), zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := middleware.GenerateToken(userID, h.secretKey)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}
