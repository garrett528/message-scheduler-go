package main

import (
	"github.com/apache/flink-statefun/statefun-sdk-go/v3/pkg/statefun"
	"github.com/go-redis/redis/v9"
)

type App struct {
	StatefunBuilder statefun.StatefulFunctions
	Redis           *redis.Client
	redisKey        string
}
