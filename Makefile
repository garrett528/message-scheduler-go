
all: compile-protos

.PHONY: compile-protos
compile-protos:
	protoc --go_out=. proto/*.proto

.PHONY: run-producer
run-producer:
	go run scripts/producer.go

.PHONY: get-scheduled-notifications
get-scheduled-notifications:
	redis-cli ZRANGE scheduled_notifications 0 -1 WITHSCORES

.PHONY: build-ingester-image
build-ingester-image:
	docker build -f Dockerfile_ingester .
