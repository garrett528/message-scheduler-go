package redisconsumer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/garrett528/message-scheduler-go/metronome/internal/utils"
	"github.com/go-redis/redis/v9"
)

type RedisConsumer struct {
	redis    *redis.Client
	redisKey string
}

func NewRedisConsumer(redisHost string, redisPort string, redisKey string) *RedisConsumer {
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	log.Printf("Redis address: %s\n", redisAddr)
	rdb := utils.NewRedisClient(redisHost, redisPort)

	return &RedisConsumer{
		redis:    rdb,
		redisKey: redisKey,
	}
}

func (rc *RedisConsumer) GetScheduledMessages(ctx context.Context, msgChan chan string) error {
	timeNow := time.Now().UnixMilli()
	zRangeArgs := &redis.ZRangeArgs{
		Key:     rc.redisKey,
		Start:   0,
		Stop:    timeNow,
		ByScore: true,
	}
	res, err := rc.redis.ZRangeArgs(ctx, *zRangeArgs).Result()
	if err != nil {
		return err
	}

	log.Printf("retrieved %d messages from redis\n", len(res))
	startScheduleTime := time.Now().UnixMilli()
	log.Printf("starting message scheduling at %d\n", startScheduleTime)
	for _, correlationId := range res {
		msgChan <- correlationId
	}
	endScheduleTime := time.Now().UnixMilli()
	log.Printf("message scheduling took %d ms", (endScheduleTime - startScheduleTime))

	return nil
}
