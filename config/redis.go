package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	ctx := context.Background()
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	log.Println("Redis connected successfully")
}
