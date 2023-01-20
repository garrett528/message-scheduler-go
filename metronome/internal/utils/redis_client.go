package utils

import (
	"fmt"

	"github.com/go-redis/redis/v9"
)

func NewRedisClient(redisHost string, redisPort string) *redis.Client {
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	fmt.Printf("Redis address: %s\n", redisAddr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}
