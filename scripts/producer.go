package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/garrett528/message-scheduler-go/proto/gen"
	"google.golang.org/protobuf/proto"
)

var (
	brokers     = flag.String("brokers", "localhost:29092", "The Kafka brokers to connect to, as a comma separated list")
	numMsg      = flag.Int("numMsg", 1000, "Number of messages to generate for testing")
	ingestTopic = flag.String("ingestTopic", "scheduled_notifications", "Ingestion topic name")
)

func main() {
	flag.Parse()

	if *brokers == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	brokerList := strings.Split(*brokers, ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))

	producer := newProducer(brokerList)
	msgs := generateMessages(*numMsg, *ingestTopic)

	for i := 0; i < len(msgs); i += 25 {
		var section []*sarama.ProducerMessage
		if i > len(msgs)-25 {
			section = msgs[i:]
		} else {
			section = msgs[i : i+25]
		}

		err := producer.SendMessages(section)
		if err != nil {
			log.Fatalf("Failed to send messages:%s", err)
		}
	}

	log.Println("Finished sending all messages")
}

func newProducer(brokerList []string) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start producer:", err)
	}

	return producer
}

func generateMessages(iterations int, topic string) []*sarama.ProducerMessage {
	var msgs []*sarama.ProducerMessage

	notificationsPerMessage := 25

	ingestRecord := gen.IngestRecord{}
	batchNum := 1
	for i := 0; i < iterations; i++ {
		correlationId := fmt.Sprintf("correlationId%d", i)
		scheduledTimeMillis := time.Now().UnixMilli()

		var messageType gen.ScheduledNotification_MessageType
		if i%2 == 0 {
			messageType = gen.ScheduledNotification_PUSH_NOTIFICATION
		} else {
			messageType = gen.ScheduledNotification_EMAIL
		}

		scheduledNotification := &gen.ScheduledNotification{
			CorrelationId:       correlationId,
			ScheduledTimeMillis: scheduledTimeMillis,
			MessageType:         messageType,
		}

		ingestRecord.ScheduledNotifications = append(ingestRecord.ScheduledNotifications, scheduledNotification)

		if (i+1)%notificationsPerMessage == 0 || i == (iterations-1) {
			ingestRecordBytes, err := proto.Marshal(&ingestRecord)
			if err != nil {
				log.Fatalln("Failed to marshal ingestRecord:", err)
			}

			batch := fmt.Sprintf("batch%d", batchNum)
			msg := &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(batch),
				Value: sarama.ByteEncoder(ingestRecordBytes),
			}

			msgs = append(msgs, msg)
			batchNum += 1
			ingestRecord = gen.IngestRecord{}
		}
	}

	return msgs
}
