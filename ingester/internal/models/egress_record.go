package models

type EgressRecord struct {
	CorrelationId       string `json:"correlationId"`
	ScheduledStatus     string `json:"scheduledStatus"`
	ScheduledTimeMillis int64  `json:"scheduledTimeMillis"`
}
