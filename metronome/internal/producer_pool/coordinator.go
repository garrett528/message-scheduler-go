package producerpool

type ProducerCoordinator struct {
	msgChan    chan string
	workerPool []*producerWorker
	topic      string
}

// set up worker pool async and begin running producer workers
// workers should be recreated if they are killed

func NewProducerCoordinator(numWorkers int, msgChan chan string, brokers string, redisHost string, redisPort string, topic string) *ProducerCoordinator {
	var workerPool []*producerWorker
	for i := 0; i < numWorkers; i++ {
		producer := newProducerWorker(brokers, redisHost, redisPort)
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
			for {
				worker.publish(pc.msgChan, pc.topic)
			}
		}(worker)
	}
}
