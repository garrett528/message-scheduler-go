package main

import (
	"flag"
	"time"

	producerpool "github.com/garrett528/message-scheduler-go/metronome/internal/producer_pool"
	redisconsumer "github.com/garrett528/message-scheduler-go/metronome/internal/redis_consumer"
	"github.com/go-co-op/gocron"
)

const CHAN_SIZE = 10000

func main() {
	redisHost := flag.String("redisHost", "redis", "Redis host")
	redisPort := flag.String("redisPort", "6379", "Redis port")
	ratePerSec := flag.Int("ratePerSec", 1000, "Sets number of messages to send to scheduled topic (per second)")
	redisKey := flag.String("redisKey", "scheduled_notifications", "Key for Redis sorted set containing scheduled payloads")
	brokers := flag.String("brokers", "localhost:29092", "The Kafka brokers to connect to, as a comma separated list")
	numWorkers := flag.Int("numWorkers", 5, "Number of producer worker goroutines to spawn")
	outTopic := flag.String("outTopic", "scheduled_out", "Name of output topic")

	flag.Parse()

	// channel to pass correlationIds from Redis consumer to message producer goroutine pool
	msgChan := make(chan string, CHAN_SIZE)

	redisConsumer := redisconsumer.NewRedisConsumer(*ratePerSec, *redisHost, *redisPort, *redisKey)
	msgProducerCoordinator := producerpool.NewProducerCoordinator(*numWorkers, msgChan, *brokers, *redisHost, *redisPort, *outTopic)

	// make sure that new runs are only scheduled after the current run is finished
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.SingletonModeAll()

	app := &App{
		Scheduler:           scheduler,
		RedisConsumer:       redisConsumer,
		ProducerCoordinator: msgProducerCoordinator,
		msgChan:             msgChan,
	}

	app.Start()
}
