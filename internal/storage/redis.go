package storage

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/internal/config"
)

func GetRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.ENV_REDIS_HOST, config.ENV_REDIS_PORT),
		Password: "", // NOTE: no password set for now
		DB:       0,  // use default DB
	})
}
