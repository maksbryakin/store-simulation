// logger/logger.go
package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

// InitLogger инициализирует логгер
func InitLogger() {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}
