package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
)

// DB представляет обёртку над database/sql.DB
type DB struct {
	*sql.DB
}

// NewDB создаёт новое подключение к базе данных PostgreSQL
func NewDB(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, nil // Если DSN не указан, возвращаем nil (БД не используется)
	}

	// Преобразуем DSN для гарантии правильного формата SSL параметров
	normalizedDSN, err := normalizeDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize DSN: %w", err)
	}

	db, err := sql.Open("postgres", normalizedDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// normalizeDSN преобразует DSN в формат, совместимый с драйвером pq
func normalizeDSN(dsn string) (string, error) {
	// Если DSN уже в формате "host=... port=...", оставляем как есть
	if strings.Contains(dsn, "host=") {
		return dsn, nil
	}

	// Парсим URL-формат DSN
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}

	// Извлекаем компоненты
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}

	user := u.User.Username()
	password, _ := u.User.Password()
	dbname := strings.TrimPrefix(u.Path, "/")

	// Парсим query параметры
	params := u.Query()
	
	// Формируем DSN в формате "key=value"
	parts := []string{
		fmt.Sprintf("host=%s", host),
		fmt.Sprintf("port=%s", port),
		fmt.Sprintf("user=%s", user),
		fmt.Sprintf("password=%s", password),
		fmt.Sprintf("dbname=%s", dbname),
	}

	// Добавляем sslmode если указан, иначе по умолчанию disable
	if sslmode := params.Get("sslmode"); sslmode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", sslmode))
	} else {
		parts = append(parts, "sslmode=disable")
	}

	// Добавляем другие параметры
	for key, values := range params {
		if key != "sslmode" && len(values) > 0 {
			parts = append(parts, fmt.Sprintf("%s=%s", key, values[0]))
		}
	}

	return strings.Join(parts, " "), nil
}

// Close закрывает соединение с базой данных
func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// Ping проверяет соединение с базой данных
func (db *DB) Ping() error {
	if db.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return db.DB.Ping()
}

