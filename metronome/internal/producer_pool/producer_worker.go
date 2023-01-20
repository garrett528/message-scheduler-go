package producerpool

import (
	"fmt"
	"log"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/garrett528/message-scheduler-go/metronome/internal/utils"
	"github.com/go-redis/redis/v9"
)

// Consists of a channel reader, Redis client, and Kafka producer
// Should get/del keys as they come in
// Should keys be batched?

type producerWorker struct {
	Producer sarama.SyncProducer
	Redis    *redis.Client
}

func newProducerWorker(brokers string, redisHost string, redisPort string) *producerWorker {
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

	return &producerWorker{
		Producer: producer,
		Redis:    rdb,
	}
}

func (pw *producerWorker) publish(msgChan chan string, topic string) error {
	correlationId := <-msgChan
	log.Printf("publishing correlationId: %s\n", correlationId)

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(correlationId),
		Value: sarama.StringEncoder(correlationId),
	}

	_, _, err := pw.Producer.SendMessage(msg)
	if err != nil {
		fmt.Printf("Failed to send messages:%s", err)
		return err
	}

	return nil
}
