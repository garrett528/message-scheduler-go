package redisconsumer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/garrett528/message-scheduler-go/metronome/internal/utils"
	"github.com/go-redis/redis/v9"
	"go.uber.org/ratelimit"
)

type RedisConsumer struct {
	Redis       *redis.Client
	RedisKey    string
	RateLimiter ratelimit.Limiter
}

func NewRedisConsumer(ratePerSec int, redisHost string, redisPort string, redisKey string) *RedisConsumer {
	log.Printf("Rate limit set to %d msg/sec\n", ratePerSec)
	rateLimiter := ratelimit.New(ratePerSec) // per second

	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	log.Printf("Redis address: %s\n", redisAddr)
	rdb := utils.NewRedisClient(redisHost, redisPort)

	return &RedisConsumer{
		Redis:       rdb,
		RedisKey:    redisKey,
		RateLimiter: rateLimiter,
	}
}

func (rc *RedisConsumer) GetScheduledMessages(ctx context.Context, msgChan chan string) error {
	timeNow := time.Now().UnixMilli()
	zRangeArgs := &redis.ZRangeArgs{
		Key:     rc.RedisKey,
		Start:   0,
		Stop:    timeNow,
		ByScore: true,
	}
	res, err := rc.Redis.ZRangeArgs(ctx, *zRangeArgs).Result()
	if err != nil {
		return err
	}

	for _, correlationId := range res {
		log.Printf("retrieved correlationId from Redis: %s\n", correlationId)
		rc.RateLimiter.Take()
		msgChan <- correlationId
	}

	return nil
}
