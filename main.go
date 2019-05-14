package main

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

func main() {
	config := api.Config{Address: "http://localhost:9090"}

	client, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("Client Error %v", err)
	}

	api := v1.NewAPI(client)

	targets, err := api.Targets(context.Background())
	fmt.Printf("Targets returned: %v\n", targets)

	query := "sum by (pod_name) (container_memory_rss{container_name=\"prometheus\"})"
	start := time.Date(2019, time.May, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2019, time.May, 15, 0, 0, 0, 0, time.UTC)
	step, _ := time.ParseDuration("1m")
	r := v1.Range{Start: start, End: end, Step: step}

	qr, err := api.QueryRange(context.Background(), query, r)
	fmt.Printf("Query returned: %+v\n", qr)
}
