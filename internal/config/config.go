package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"os"
)

var (
	ErrEmptyServerAddress  = errors.New("server address is empty")
	ErrSecretKeyGeneration = errors.New("failed to generate secret key")
)

type Config struct {
	// Адрес и порт запуска сервиса
	RunAddress string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`

	// Адрес подключения к базе данных PostgreSQL
	DatabaseURI string `env:"DATABASE_URI"`

	// Адрес системы расчёта начислений
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	// Уровень логирования
	LogLevel string `env:"LOG_LEVEL" envDefault:"FATAL"`

	// Секретный ключ для подписи кук
	SecretKey string `env:"SECRET_KEY"`
}

var (
	flagRunAddress           = flag.String("a", "localhost:8080", "Адрес и порт запуска сервиса")
	flagDatabaseURI          = flag.String("d", "", "Адрес подключения к базе данных")
	flagAccrualSystemAddress = flag.String("r", "", "Адрес системы расчёта начислений")
	flagLogLevel             = flag.String("l", "FATAL", "Уровень логирования")
	flagSecretKey            = flag.String("s", "", "Секретный ключ для подписи кук")
)

func NewConfig() (*Config, error) {
	// Парсим флаги только если они ещё не распарсены
	if !flag.Parsed() {
		flag.Parse()
	}

	cfg := &Config{
		RunAddress:           *flagRunAddress,
		DatabaseURI:          *flagDatabaseURI,
		AccrualSystemAddress: *flagAccrualSystemAddress,
		LogLevel:             *flagLogLevel,
		SecretKey:            *flagSecretKey,
	}

	// Переменные окружения перезаписывают флаги, если они установлены
	if val, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		cfg.RunAddress = val
	}
	if val, ok := os.LookupEnv("DATABASE_URI"); ok {
		cfg.DatabaseURI = val
	}
	if val, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		cfg.AccrualSystemAddress = val
	}
	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = val
	}

	if val, ok := os.LookupEnv("SECRET_KEY"); ok {
		cfg.SecretKey = val
	}

	// Если секретный ключ не задан, генерируем случайный
	if cfg.SecretKey == "" {
		secretKey, err := generateSecretKey()
		if err != nil {
			return nil, ErrSecretKeyGeneration
		}
		cfg.SecretKey = secretKey
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.RunAddress == "" {
		return ErrEmptyServerAddress
	}

	return nil
}

// generateSecretKey генерирует случайный секретный ключ
func generateSecretKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
