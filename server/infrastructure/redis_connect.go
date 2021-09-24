package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	client *redis.Client
)

func loadRedis() {
	address := fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port)
	password := config.Redis.Password
	db := config.Redis.Db
	client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Fatal("Unable to connect redis: ", err)
	}
}

func GetRedisClient() *redis.Client {
	return client
}
