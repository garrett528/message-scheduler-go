package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	producerpool "github.com/garrett528/message-scheduler-go/metronome/internal/producer_pool"
	redisconsumer "github.com/garrett528/message-scheduler-go/metronome/internal/redis_consumer"
	"github.com/go-co-op/gocron"
)

type App struct {
	Scheduler           *gocron.Scheduler
	RedisConsumer       *redisconsumer.RedisConsumer
	ProducerCoordinator *producerpool.ProducerCoordinator
	msgChan             chan string
}

func (app *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			signal.Stop(sigs)
			cancel()
		case <-ctx.Done():
			// exit gracefully
		}
	}()

	// start cron to consume from redis every minute
	_, _ = app.Scheduler.Every(60).Seconds().Do(app.RedisConsumer.GetScheduledMessages, ctx, app.msgChan)
	app.Scheduler.StartAsync()

	// start message producers to read from channel + produce messages
	app.ProducerCoordinator.StartAsync()

	fmt.Println("service started...")
	<-ctx.Done()
	fmt.Println("\nexiting...")
}
