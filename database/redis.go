package database

import (
	"WebCrawler/config"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

func GetRedisClient(cfg *config.Config) *redis.Client {
	redisOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddress,
			Password: cfg.RedisPass,
			DB:       0,
		})
	})
	return redisClient
}
