// cache/cache.go
package cache

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/maksbryakin/store-simulation/logger"
	"github.com/maksbryakin/store-simulation/models"
	"go.uber.org/zap"
)

var ctx = context.Background()

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // Нет пароля по умолчанию
		DB:       0,  // Используем DB 0
	})
	return &RedisCache{Client: client}
}

func (rc *RedisCache) SetProducts(products []models.Product) error {
	data, err := json.Marshal(products)
	if err != nil {
		logger.Logger.Error("Ошибка при маршалинге продуктов", zap.Error(err))
		return err
	}
	err = rc.Client.Set(ctx, "products", data, 0).Err()
	if err != nil {
		logger.Logger.Error("Ошибка при установке продуктов в Redis", zap.Error(err))
	}
	return err
}

func (rc *RedisCache) GetProducts() ([]models.Product, error) {
	val, err := rc.Client.Get(ctx, "products").Result()
	if err != nil {
		logger.Logger.Error("Ошибка при получении продуктов из Redis", zap.Error(err))
		return nil, err
	}
	var products []models.Product
	err = json.Unmarshal([]byte(val), &products)
	if err != nil {
		logger.Logger.Error("Ошибка при анмаршалинге продуктов из Redis", zap.Error(err))
		return nil, err
	}
	return products, nil
}
