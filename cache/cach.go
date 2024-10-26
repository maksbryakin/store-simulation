// cache/cache.go
package cache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/maksbryakin/store-simulation/models"
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
		return err
	}
	return rc.Client.Set(ctx, "products", data, 0).Err()
}

func (rc *RedisCache) GetProducts() ([]models.Product, error) {
	val, err := rc.Client.Get(ctx, "products").Result()
	if err != nil {
		return nil, err
	}
	var products []models.Product
	err = json.Unmarshal([]byte(val), &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}
