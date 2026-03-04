package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/mmeow0/gophermart-bonus/internal/middleware"
	"github.com/mmeow0/gophermart-bonus/internal/model"
	"github.com/mmeow0/gophermart-bonus/internal/repository"
	"github.com/mmeow0/gophermart-bonus/internal/service"
	"github.com/mmeow0/gophermart-bonus/internal/util"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderService *service.OrderService
	logger       *zap.Logger
}

func NewOrderHandler(orderService *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger,
	}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !util.IsValidLuhn(orderNumber) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = h.orderService.CreateOrder(r.Context(), orderNumber, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, repository.ErrOrderExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.logger.Error("Failed to create order", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		h.logger.Error("Failed to encode orders", zap.Error(err))
	}
}

func (h *OrderHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	current, withdrawn, err := h.orderService.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get balance", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := model.BalanceResponse{
		Current:   current,
		Withdrawn: withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode balance", zap.Error(err))
	}
}

func (h *OrderHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req model.WithdrawRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Order == "" || req.Sum <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !util.IsValidLuhn(req.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err := h.orderService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		h.logger.Error("Failed to withdraw", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.orderService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get withdrawals", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Error("Failed to encode withdrawals", zap.Error(err))
	}
}
