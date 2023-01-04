#!/usr/bin/env bash
curl -X POST \
     -H "Content-Type: application/vnd.kafka.json.v2+json" \
     -H "Accept: application/vnd.kafka.v2+json" \
     --data '{"records":[{"key":"correlation1","value":{"correlationId":"correlation1","messageType":1,"utcHour":3}}]}' \
     "http://localhost:8082/topics/scheduled_notifications"
