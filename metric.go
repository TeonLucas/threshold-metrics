package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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
type Attributes map[string]any

func (data *AccountData) count(timeslice interface{}) (count float64) {
	aggregate := timeslice.(map[string]interface{})
	count = aggregate["count"].(float64)
	//log.Printf("DEBUG count=%f", count)
	return
}

func (data *AccountData) countBelow(timeslice interface{}) (countAbove float64) {
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
	area := percentage
	countAbove = area * count
	if countAbove < 0 {
		countAbove = 0
	}
	//log.Printf("DEBUG countAbove=%f count=%f mean=%f threshold=%f, timeslice=%+v", countAbove, count, mean, data.Threshold, timeslice)
	return
}

func (data *AccountData) os(attributes Attributes) (os string) {
	var appName = attributes["appName"].(string)
	os = ""
	if strings.Contains(appName, "android") {
		os = "android"
	} else if strings.Contains(appName, "ios") {
		os = "ios"
	}
	return
}

func (data *AccountData) pushMetric(timestamp int64, timeslice interface{}, attributes Attributes) {
	var metric Metric
	metric.Name = data.NewMetricName
	metric.Type = "gauge"
	metric.Value = data.countBelow(timeslice)
	metric.Timestamp = timestamp
	attributes["TotalCount"] = data.count(timeslice)
	attributes["os"] = data.os(attributes)
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

		// log.Printf("DEBUG metrics: %s", j)

		b := retryQuery(data.Client, "POST", MetricEndoint, string(j), data.MetricHeaders)
		log.Printf("Submitted %s", b)
	}

	// Clear metrics that were sent
	data.Metrics = nil
}

func (data *AccountData) makeMetricsToCSV() {
	// var err error

	// Send array of metrics to api
	if len(data.Metrics) == 0 {
		log.Println("No metrics to send")
	} else {
		filePath := "metrics_" + strconv.FormatFloat(data.Threshold, 'f', -1, 64) + ".csv" // Convert data.Threshold to string before concatenating
		err := saveCSV(filePath, data)
		if err != nil {
			log.Printf("Error exporting metrics to host: %v", err)
		} else {
			log.Printf("Saved %d metrics to CSV on the host machine", len(data.Metrics))
		}
	}

	// Clear metrics that were sent
	data.Metrics = nil
}

func saveCSV(filePath string, data *AccountData) error {
	hostFilePath, _ := filepath.Abs(filePath)
	dirPath := path.Dir(hostFilePath)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	file, err := os.OpenFile(hostFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"Name",
		"Type",
		"Count Below",
		"Total Count",
		"Threshold",
		"OS",
		"Timestamp",
		"AppName",
		"Attributes",
		// Add more headers as needed
	}
	err = writer.Write(headers)
	if err != nil {
		return fmt.Errorf("error writing headers to CSV file: %v", err)
	}

	for _, metric := range data.Metrics {
		record := []string{
			metric.Name,
			metric.Type,
			fmt.Sprintf("%f", metric.Value),
			fmt.Sprintf("%f", metric.Attributes["TotalCount"]),
			fmt.Sprintf("%f", data.Threshold),
			fmt.Sprintf("%v", metric.Attributes["os"]),
			fmt.Sprintf("%d", metric.Timestamp),
			fmt.Sprintf("%v", metric.Attributes["appName"]),
			fmt.Sprintf("%v", metric.Attributes),
			// Add more fields as needed
		}
		err := writer.Write(record)
		if err != nil {
			return fmt.Errorf("error writing record to CSV file: %v", err)
		}
	}

	return nil
}
