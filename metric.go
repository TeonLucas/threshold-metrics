package main

import (
	"encoding/json"
	"log"
	"math"
)

const (
	MetricEndoint = "https://metric-api.newrelic.com/metric/v1"
)

type MetricPayload struct {
	Metrics []Metric `json:"metrics"`
}

type Metric struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Value      float64    `json:"value"`
	Timestamp  int64      `json:"timestamp"`
	Attributes Attributes `json:"attributes"`
}
type Attributes map[string]string

func (data *AccountData) countAbove(timeslice interface{}) (countAbove float64) {
	aggregate := timeslice.(map[string]interface{})
	count := aggregate["count"].(float64)
	if count == 0.0 {
		return 0
	}
	mean := aggregate["total"].(float64) / count
	if mean == 0 {
		//log.Printf("DEBUG count=%f mean=%f countAbove=%f", count, 0.0, 0.0)
		return 0
	}
	std := math.Sqrt(aggregate["sumOfSquares"].(float64) / count)
	zscore := (data.Threshold - mean) / std
	var percentage float64
	if zscore > 4 {
		percentage = 1.0
	} else {
		percentage = data.ZTable.FindPercentage(zscore)
	}
	area := 1 - percentage
	countAbove = area * count
	if countAbove < 0 {
		countAbove = 0
	}
	//log.Printf("DEBUG countAbove=%f count=%f mean=%f threshold=%f, timeslice=%+v", countAbove, count, mean, data.Threshold, timeslice)
	return
}

func (data *AccountData) pushMetric(timestamp int64, timeslice interface{}, attributes Attributes) {
	var metric Metric
	metric.Name = data.NewMetricName
	metric.Type = "gauge"
	metric.Value = data.countAbove(timeslice)
	metric.Timestamp = timestamp
	metric.Attributes = attributes
	data.Metrics = append(data.Metrics, metric)
}

func (data *AccountData) makeMetrics() {
	var err error
	var j []byte

	// Send array of metrics to api
	if len(data.Metrics) == 0 {
		log.Println("No metrics to send")
	} else {
		j, err = json.Marshal([]MetricPayload{{Metrics: data.Metrics}})
		if err != nil {
			log.Printf("Error creating Metric payload: %v", err)
		}
		log.Printf("Sending %d metrics to the metric api", len(data.Metrics))

		//log.Printf("DEBUG metrics: %s", j)

		b := retryQuery(data.Client, "POST", MetricEndoint, string(j), data.MetricHeaders)
		log.Printf("Submitted %s", b)
	}

	// Clear metrics that were sent
	data.Metrics = nil
}
