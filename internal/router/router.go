package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/mmeow0/gophermart-bonus/internal/handler"
	"github.com/mmeow0/gophermart-bonus/internal/logger"
	"github.com/mmeow0/gophermart-bonus/internal/middleware"
	"go.uber.org/zap"
)

func NewRouter(
	userHandler *handler.UserHandler,
	orderHandler *handler.OrderHandler,
	pingHandler *handler.PingHandler,
	secretKey string,
	log *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(logger.RequestLogger(log))
	r.Use(middleware.AuthMiddleware(secretKey, log))

	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)
	r.Post("/api/user/orders", orderHandler.UploadOrder)
	r.Get("/api/user/orders", orderHandler.GetOrders)
	r.Get("/api/user/balance", orderHandler.GetBalance)
	r.Post("/api/user/balance/withdraw", orderHandler.Withdraw)
	r.Get("/api/user/withdrawals", orderHandler.GetWithdrawals)
	r.Get("/ping", pingHandler.Ping)

	return r
}
