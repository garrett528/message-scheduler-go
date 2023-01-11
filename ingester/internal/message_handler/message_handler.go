package message_handler

import (
	"context"
	"encoding/json"

	"github.com/garrett528/message-scheduler-go/ingester/internal/models"
	"github.com/garrett528/message-scheduler-go/proto/gen"
	"github.com/go-redis/redis/v9"
)

func HandleMessage(ctx context.Context, record *gen.IngestRecord, client *redis.Client) []*models.EgressRecord {
	var egressRecords []*models.EgressRecord
	pipe := client.Pipeline()

	for _, notification := range record.ScheduledNotifications {
		notificationJson, err := json.Marshal(notification)
		if err != nil {
			egressRecord := &models.EgressRecord{
				CorrelationId:       notification.CorrelationId,
				ScheduledStatus:     "failed",
				ScheduledTimeMillis: notification.ScheduledTimeMillis,
			}
			egressRecords = append(egressRecords, egressRecord)
			continue
		}

		pipe.ZAdd(ctx, "scheduled_notifications", redis.Z{
			Score:  float64(notification.ScheduledTimeMillis),
			Member: notificationJson,
		})

		egressRecord := &models.EgressRecord{
			CorrelationId:       notification.CorrelationId,
			ScheduledStatus:     "scheduled",
			ScheduledTimeMillis: notification.ScheduledTimeMillis,
		}
		egressRecords = append(egressRecords, egressRecord)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		panic(err)
	}

	return egressRecords
}
