package main

import (
	"fmt"

	"github.com/apache/flink-statefun/statefun-sdk-go/v3/pkg/statefun"
	"github.com/garrett528/message-scheduler-go/ingester/internal/message_handler"
	"github.com/garrett528/message-scheduler-go/ingester/internal/models"
	"github.com/garrett528/message-scheduler-go/proto/gen"
)

var (
	IngestRecordType         = statefun.MakeProtobufType(&gen.IngestRecord{})
	EgressRecordType         = statefun.MakeJsonType(statefun.TypeNameFrom("message-scheduler-go/EgressRecord"))
	IngesterStatefunTypeName = statefun.TypeNameFrom("message-scheduler-go/IngesterType")
	KafkaEgressTypeName      = statefun.TypeNameFrom("message-scheduler-go/IngesterEgress")
)

func (app *App) Ingest(ctx statefun.Context, message statefun.Message) error {
	var ingestRecord gen.IngestRecord
	if err := message.As(IngestRecordType, &ingestRecord); err != nil {
		return fmt.Errorf("failed to deserialize ingest record: %w", err)
	}

	scheduledStatus := "scheduled"
	if err := message_handler.HandleMessage(ctx, &ingestRecord, app.Redis); err != nil {
		scheduledStatus = "failed"
	}

	egressRecord := models.EgressRecord{
		ScheduledStatus: scheduledStatus,
	}

	ctx.SendEgress(&statefun.KafkaEgressBuilder{
		Target:    KafkaEgressTypeName,
		Topic:     "ingest_out",
		Key:       ingestRecord.CorrelationId,
		Value:     egressRecord,
		ValueType: EgressRecordType,
	})

	return nil
}
