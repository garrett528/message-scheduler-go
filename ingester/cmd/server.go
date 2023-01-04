package main

import (
	"fmt"
	"net/http"

	"github.com/apache/flink-statefun/statefun-sdk-go/v3/pkg/statefun"
)

type IngestRequest struct {
	CorrelationId string `json:"correlationId"`
	MessageType   int32  `json:"messageType"`
	UtcHour       int32  `json:"utcHour"`
}

type EgressRecord struct {
	Payload string `json:"payload"`
}

var (
	IngestRecordType         = statefun.MakeJsonType(statefun.TypeNameFrom("message-scheduler-go/IngestRequest"))
	EgressRecordType         = statefun.MakeJsonType(statefun.TypeNameFrom("message-scheduler-go/EgressRecord"))
	IngesterStatefunTypeName = statefun.TypeNameFrom("message-scheduler-go/IngesterType")
	KafkaEgressTypeName      = statefun.TypeNameFrom("message-scheduler-go/IngesterEgress")
)

func Ingest(ctx statefun.Context, message statefun.Message) error {
	var request IngestRequest
	if err := message.As(IngestRecordType, &request); err != nil {
		return fmt.Errorf("failed to deserialize ingest request: %w", err)
	}

	schedule_out := fmt.Sprintf("correlationId: %s | messageType: %d | utcHour: %d", request.CorrelationId, request.MessageType, request.UtcHour)

	egressRecord := EgressRecord{
		Payload: schedule_out,
	}

	ctx.SendEgress(&statefun.KafkaEgressBuilder{
		Target:    KafkaEgressTypeName,
		Topic:     "ingest_out",
		Key:       request.CorrelationId,
		Value:     egressRecord,
		ValueType: EgressRecordType,
	})

	return nil
}

func main() {

	builder := statefun.StatefulFunctionsBuilder()

	_ = builder.WithSpec(statefun.StatefulFunctionSpec{
		FunctionType: IngesterStatefunTypeName,
		Function:     statefun.StatefulFunctionPointer(Ingest),
	})

	http.Handle("/statefun", builder.AsHandler())
	_ = http.ListenAndServe(":8000", nil)
}
