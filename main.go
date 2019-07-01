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

var insecureTLSFlag bool
var durationFlag int
var stepFlag, tokenFlag, urlFlag string

func initFlags() {
	flag.BoolVar(&insecureTLSFlag, "insecure", false, "Trust self-signed HTTP certificates")
	flag.IntVar(&durationFlag, "duration", 30, "Duration of test in integer minutes (used to calculate quest start time)")
	flag.StringVar(&stepFlag, "step", "1m", "Query resolution step width in number of seconds")
	flag.StringVar(&tokenFlag, "token", "", "Authorization type + token for endpoint")
	flag.StringVar(&urlFlag, "url", "http://localhost:9090", "URL for prometheus connection")
	flag.Parse()
}

func main() {
	initFlags()

	// Check if no flags were passed, print help
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		return
	}

	config := api.Config{Address: urlFlag, Authorization: tokenFlag, InsecureTLS: insecureTLSFlag}
	client, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("Client Error %v\n", err)
		return
	}

	api := v1.NewAPI(client)
	query := "sum by (pod_name) (container_memory_rss{container_name=\"prometheus\"})"
	end := time.Now()
	start := end.Add(time.Duration(-1 * durationFlag) * time.Minute)
	step, err := time.ParseDuration("1m")
	if err != nil {
		fmt.Printf("Error parsing step duration %s\n", stepFlag)
	}
	r := v1.Range{Start: start, End: end, Step: step}

	queryResult, err := api.QueryRange(context.Background(), query, r)
	if err != nil {
		fmt.Printf("Query Error: %v\n", err)
		return
	}
	fmt.Printf("Query returned: %+v\n", queryResult)

	data, ok := queryResult.(model.Matrix)
	if !ok {
		fmt.Printf("Unsupported result format: %s\n", queryResult.Type().String())
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
