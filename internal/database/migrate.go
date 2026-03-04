package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations применяет все миграции к базе данных используя golang-migrate
func RunMigrations(db *sql.DB, migrationsPath string) error {
	// Ищем папку с миграциями
	actualPath, err := findMigrationsPath(migrationsPath)
	if err != nil {
		return fmt.Errorf("migrations directory not found at path '%s': please ensure migrations exist before starting the application", migrationsPath)
	}

	// Создаём драйвер для PostgreSQL
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Создаём экземпляр migrate
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+actualPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations from '%s': %w", actualPath, err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// findMigrationsPath ищет папку с миграциями
func findMigrationsPath(basePath string) (string, error) {
	// Проверяем возможные пути
	paths := []string{
		basePath,
		filepath.Join(".", basePath),
		filepath.Join("..", basePath),
		filepath.Join("../..", basePath),
		filepath.Join("../../..", basePath),
	}

	// Получаем текущую рабочую директорию
	wd, _ := os.Getwd()
	paths = append(paths, filepath.Join(wd, basePath))

	// Проверяем путь относительно исполняемого файла
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		paths = append(paths, filepath.Join(exeDir, basePath))
		paths = append(paths, filepath.Join(exeDir, "..", basePath))
	}

	for _, path := range paths {
		if absPath, err := filepath.Abs(path); err == nil {
			if stat, err := os.Stat(absPath); err == nil && stat.IsDir() {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("migrations directory not found")
}

