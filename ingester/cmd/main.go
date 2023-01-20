package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/apache/flink-statefun/statefun-sdk-go/v3/pkg/statefun"
	"github.com/go-redis/redis/v9"
)

func main() {
	redisHost := flag.String("redisHost", "redis", "Redis host")
	redisPort := flag.String("redisPort", "6379", "Redis port")

	flag.Parse()

	redisAddr := fmt.Sprintf("%s:%s", *redisHost, *redisPort)
	fmt.Printf("Redis address: %s\n", redisAddr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	app := &App{
		StatefunBuilder: statefun.StatefulFunctionsBuilder(),
		Redis:           rdb,
	}

	_ = app.StatefunBuilder.WithSpec(statefun.StatefulFunctionSpec{
		FunctionType: IngesterStatefunTypeName,
		Function:     statefun.StatefulFunctionPointer(app.Ingest),
	})

	http.Handle("/statefun", app.StatefunBuilder.AsHandler())
	_ = http.ListenAndServe(":8000", nil)
}
