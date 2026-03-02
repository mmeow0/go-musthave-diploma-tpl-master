package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mmeow0/gophermart-bonus/internal/client"
	"github.com/mmeow0/gophermart-bonus/internal/config"
	"github.com/mmeow0/gophermart-bonus/internal/database"
	"github.com/mmeow0/gophermart-bonus/internal/handler"
	"github.com/mmeow0/gophermart-bonus/internal/logger"
	"github.com/mmeow0/gophermart-bonus/internal/repository"
	"github.com/mmeow0/gophermart-bonus/internal/router"
	"github.com/mmeow0/gophermart-bonus/internal/service"
	"github.com/mmeow0/gophermart-bonus/internal/worker"
	"go.uber.org/zap"
)

type App struct {
	cfg           *config.Config
	router        http.Handler
	db            *database.DB
	logger        *zap.Logger
	accrualWorker *worker.AccrualWorker
	ctx           context.Context
	cancel        context.CancelFunc
}

func InitializeApp() (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	if cfg.DatabaseURI == "" {
		return nil, fmt.Errorf("database URI is required")
	}

	db, err := database.NewDB(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := database.RunMigrations(db.DB, "migrations"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	userRepo := repository.NewPostgresUserRepository(db.DB)
	orderRepo := repository.NewPostgresOrderRepository(db.DB)
	balanceRepo := repository.NewPostgresBalanceRepository(db.DB)

	userService := service.NewUserService(userRepo)
	orderService := service.NewOrderService(orderRepo, balanceRepo)

	userHandler := handler.NewUserHandler(userService, cfg.SecretKey, log)
	orderHandler := handler.NewOrderHandler(orderService, log)
	pingHandler := handler.NewPingHandler(db, log)

	rt := router.NewRouter(userHandler, orderHandler, pingHandler, cfg.SecretKey, log)

	ctx, cancel := context.WithCancel(context.Background())

	var accrualWorker *worker.AccrualWorker
	if cfg.AccrualSystemAddress != "" {
		accrualClient := client.NewAccrualClient(cfg.AccrualSystemAddress)
		accrualWorker = worker.NewAccrualWorker(orderService, accrualClient, log)
		accrualWorker.Start(ctx)
		log.Info("Accrual worker started", zap.String("accrual_address", cfg.AccrualSystemAddress))
	} else {
		log.Warn("Accrual system address not configured, worker will not start")
	}

	log.Info("Using PostgreSQL storage", zap.String("dsn", maskDSN(cfg.DatabaseURI)))

	return &App{
		cfg:           cfg,
		router:        rt,
		db:            db,
		logger:        log,
		accrualWorker: accrualWorker,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

func maskDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return "***"
	}

	if u.User != nil {
		username := u.User.Username()
		u.User = url.UserPassword(username, "***")
	}

	return u.String()
}

func (a *App) Run() error {
	a.logger.Info("Starting server", zap.String("address", a.cfg.RunAddress))
	return http.ListenAndServe(a.cfg.RunAddress, a.router)
}

func (a *App) Close() error {
	if a.cancel != nil {
		a.cancel()
	}

	if a.accrualWorker != nil {
		a.accrualWorker.Stop()
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Error("Failed to close database", zap.Error(err))
		}
	}

	return nil
}
