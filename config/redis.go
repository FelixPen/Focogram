package config

import (
	"Focogram/global"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func initRedis() {
	RedisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	})
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to connect redis, got error: %v", err)
	}
	global.Redis = RedisClient
}
