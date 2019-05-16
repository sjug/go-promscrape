package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type TimePoint [2]float64
type TimeSeriesPoints []TimePoint

type TimeSeries struct {
	Name   string
	Points TimeSeriesPoints
	Tags   map[string]string
	Values []float64
}

func main() {
	urlFlag := flag.String("url", "http://localhost:9090", "URL for prometheus connection (ie. http://localhost:9090)")
	flag.Parse()

	config := api.Config{Address: *urlFlag}

	client, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("Client Error %v", err)
		return
	}

	api := v1.NewAPI(client)

	query := "sum by (pod_name) (container_memory_rss{container_name=\"prometheus\"})"
	start := time.Date(2019, time.May, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2019, time.May, 15, 0, 0, 0, 0, time.UTC)
	step, _ := time.ParseDuration("1m")
	r := v1.Range{Start: start, End: end, Step: step}

	queryResult, err := api.QueryRange(context.Background(), query, r)
	if err != nil {
		fmt.Printf("Query Error: %v", err)
		return
	}
	fmt.Printf("Query returned: %+v\n", queryResult)

	data, ok := queryResult.(model.Matrix)
	if !ok {
		fmt.Errorf("Unsupported result format: %s", queryResult.Type().String())
		return
	}

	series := TimeSeries{
		Name: query,
		Tags: map[string]string{},
	}

	for k, v := range data[0].Metric {
		series.Tags[string(k)] = string(v)
	}

	for _, v := range data[0].Values {
		series.Points = append(series.Points, TimePoint{float64(v.Value), float64(v.Timestamp.Unix() * 1000)})
		series.Values = append(series.Values, float64(v.Value))
	}

	fmt.Printf("Series: %+v\n", series)

}
