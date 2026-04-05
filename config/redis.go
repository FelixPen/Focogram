package config

import (
	"Focogram/global"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func initRedis() {
	addr := "127.0.0.1:6379"
	password := ""
	db := 0

	if AppConfig.Redis.Addr != "" {
		addr = AppConfig.Redis.Addr
	}
	if AppConfig.Redis.Password != "" {
		password = AppConfig.Redis.Password
	}
	if AppConfig.Redis.DB > 0 {
		db = AppConfig.Redis.DB
	}

	RedisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to connect redis, got error: %v", err)
	}
	global.Redis = RedisClient
}
