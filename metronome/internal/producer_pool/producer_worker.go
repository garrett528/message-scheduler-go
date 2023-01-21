package producerpool

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/garrett528/message-scheduler-go/metronome/internal/utils"
	"github.com/go-redis/redis/v9"
)

const MSG_BUFFER_SIZE = 1000

type producerWorker struct {
	producer    sarama.SyncProducer
	redis       *redis.Client
	redisKey    string
	ctx         context.Context
	timeStart   int64
	timeEnd     int64
	prevTimeEnd int64
	publishEnd  bool
}

func newProducerWorker(brokers string, redisHost string, redisPort string, redisKey string) *producerWorker {
	brokerList := strings.Split(brokers, ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start producer:", err)
	}
	rdb := utils.NewRedisClient(redisHost, redisPort)

	timeNow := time.Now().UTC().UnixMilli()

	return &producerWorker{
		producer:    producer,
		redis:       rdb,
		redisKey:    redisKey,
		ctx:         context.Background(),
		timeStart:   timeNow,
		timeEnd:     timeNow,
		prevTimeEnd: timeNow,
		publishEnd:  true,
	}
}

func (pw *producerWorker) publish(msgChan chan string, topic string) error {
	var idBuf []string
	for i := 0; i < MSG_BUFFER_SIZE; i++ {
		if len(msgChan) > 0 {
			correlationId := <-msgChan
			idBuf = append(idBuf, correlationId)
		} else {
			break
		}
	}

	res, err := pw.redis.MGet(pw.ctx, idBuf...).Result()
	if err != nil {
		return err
	}

	msgsBuf := make([]*sarama.ProducerMessage, len(res))
	for i, msgData := range res {
		producerMsg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(idBuf[i]),
			Value: sarama.StringEncoder(msgData.(string)),
		}

		msgsBuf[i] = producerMsg
	}

	pw.producer.SendMessages(msgsBuf)

	// remove the entries from Redis after processed
	pipe := pw.redis.Pipeline()
	pipe.ZRem(pw.ctx, pw.redisKey, idBuf)
	pipe.Del(pw.ctx, idBuf...)
	_, err = pipe.Exec(pw.ctx)
	if err != nil {
		return err
	}

	pw.timeEnd = time.Now().UnixMilli()

	return nil
}

func (pw *producerWorker) publishTimeEnd() {
	if pw.timeEnd != pw.timeStart && pw.prevTimeEnd == pw.timeEnd && pw.publishEnd {
		log.Printf("publishing took %d ms", pw.timeEnd-pw.timeStart)
		pw.publishEnd = false
	} else {
		pw.prevTimeEnd = pw.timeEnd
	}
}

func (pw *producerWorker) close() error {
	err := pw.producer.Close()
	if err != nil {
		return err
	}

	return nil
}
