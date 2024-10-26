// database/database.go
package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maksbryakin/store-simulation/logger"
)

// DB глобальная переменная для доступа к базе данных
var DB *pgxpool.Pool

// InitDB инициализирует соединение с базой данных
func InitDB() error {
	// Получение строки подключения из переменных окружения
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:12345qwe@localhost:5432/store_simulation"
	}

	var err error
	DB, err = pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных: %v", err)
	}

	logger.Logger.Info("Подключение к базе данных успешно")
	return nil
}
