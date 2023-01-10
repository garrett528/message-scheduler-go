package message_handler

import (
	"context"
	"encoding/json"

	"github.com/garrett528/message-scheduler-go/proto/gen"
	"github.com/go-redis/redis/v9"
)

func HandleMessage(ctx context.Context, record *gen.IngestRecord, client *redis.Client) error {
	pipe := client.Pipeline()

	recordBin, err := json.Marshal(record)
	if err != nil {
		panic(err)
	}

	pipe.ZAdd(ctx, "scheduled_notifications", redis.Z{
		Score:  float64(record.ScheduledTimeMillis),
		Member: recordBin,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		panic(err)
	}

	return nil
}
