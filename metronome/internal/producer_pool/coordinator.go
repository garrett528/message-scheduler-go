package producerpool

import (
	"time"
)

type ProducerCoordinator struct {
	msgChan    chan string
	workerPool []*producerWorker
	topic      string
}

// workers should be recreated if they are killed

func NewProducerCoordinator(numWorkers int, msgChan chan string, brokers string, redisHost string, redisPort string, redisKey string, topic string) *ProducerCoordinator {
	var workerPool []*producerWorker
	for i := 0; i < numWorkers; i++ {
		producer := newProducerWorker(brokers, redisHost, redisPort, redisKey)
		workerPool = append(workerPool, producer)
	}

	return &ProducerCoordinator{
		msgChan:    msgChan,
		workerPool: workerPool,
		topic:      topic,
	}
}

func (pc *ProducerCoordinator) StartAsync() {
	for _, worker := range pc.workerPool {
		go func(worker *producerWorker) {
			defer worker.close()
			for {
				// only begin publishing if there are messages to publish
				// otherwise we waste calls to Redis
				if len(pc.msgChan) > 0 {
					worker.publish(pc.msgChan, pc.topic)
				} else {
					worker.publishTimeEnd()
					time.Sleep(time.Duration(time.Duration(10).Milliseconds()))
				}
			}
		}(worker)
	}
}
