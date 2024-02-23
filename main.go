package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/iancoleman/strcase"
)

const (
	DefaultPollInterval = "1m"
)

// To store and analyze data
type AccountData struct {
	AccountId      string
	MetricName     string
	NewMetricName  string
	MetricWhere    string
	MetricFacet    string
	Attributes     []string
	Metrics        []Metric
	Timestamp      int64
	Threshold      float64
	LicenseKey     string
	UserKey        string
	Client         *http.Client
	GraphQlHeaders []string
	MetricHeaders  []string
	Details        Details
	SampleTime     int64
	PollInterval   time.Duration
}

type Details struct {
	EntityGuid  string `json:"entityGuid"`
	CurrentTime int64  `json:"-"`
	StartTime   int64  `json:"startTimeMs"`
	Duration    int    `json:"durationMs"`
}

func (data *AccountData) makeClient() {
	data.Client = &http.Client{}
	data.GraphQlHeaders = []string{"Content-Type:application/json", "API-Key:" + data.UserKey}
	data.MetricHeaders = []string{"Content-Type:application/json", "Api-Key:" + data.LicenseKey}
	// make list of attributes for each metric
	attributes := strings.Split(data.MetricFacet, ",")
	for _, attributeRaw := range attributes {
		attribute := strings.TrimSpace(attributeRaw)
		if attribute == "entity.guid" || attribute == "entityGuid" {
			continue
		}
		data.Attributes = append(data.Attributes, attribute)
	}
	data.Attributes = append(data.Attributes, "entity.guid")
}

func main() {
	var err error

	// Get required settings
	data := AccountData{
		AccountId:   strings.TrimSpace(os.Getenv("NEW_RELIC_ACCOUNT")),
		MetricName:  strings.TrimSpace(os.Getenv("METRIC_NAME")),
		MetricWhere: strings.TrimSpace(os.Getenv("METRIC_WHERE")),
		MetricFacet: strings.TrimSpace(os.Getenv("METRIC_FACET")),
		LicenseKey:  strings.TrimSpace(os.Getenv("NEW_RELIC_LICENSE_KEY")),
		UserKey:     strings.TrimSpace(os.Getenv("NEW_RELIC_USER_KEY")),
	}
	if len(data.AccountId) == 0 {
		log.Printf("Please set env var NEW_RELIC_ACCOUNT")
		os.Exit(0)
	}
	if len(data.MetricName) == 0 {
		log.Printf("Please set env var METRIC_NAME")
		os.Exit(0)
	}
	data.NewMetricName = strcase.ToLowerCamel(data.MetricName) + "Threshold"
	if len(data.MetricWhere) == 0 {
		log.Printf("Please set env var METRIC_WHERE")
		os.Exit(0)
	}
	if len(data.MetricFacet) == 0 {
		log.Printf("Please set env var METRIC_FACET")
		os.Exit(0)
	}
	threshold := os.Getenv("THRESHOLD")
	if len(threshold) == 0 {
		log.Printf("Please set env var THRESHOLD")
		os.Exit(0)
	}
	data.Threshold, err = strconv.ParseFloat(threshold, 64)
	if err != nil {
		log.Printf("Invalid number for env var THRESHOLD: %v", err)
		os.Exit(0)
	}
	if len(data.LicenseKey) == 0 {
		log.Printf("Please set env var NEW_RELIC_LICENSE_KEY")
		os.Exit(0)
	}
	if len(data.UserKey) == 0 {
		log.Printf("Please set env var NEW_RELIC_USER_KEY")
		os.Exit(0)
	}

	// Get poll interval
	pollInterval := os.Getenv("POLL_INTERVAL")
	if len(pollInterval) == 0 {
		pollInterval = DefaultPollInterval
	}
	data.PollInterval, err = time.ParseDuration(pollInterval)
	if err != nil {
		log.Fatalf("Error: could not parse env var POLL_INTERVAL: %s, must be a duration (ex: 1h)", err)
	}
	if data.PollInterval < time.Minute {
		log.Fatalf("Error: POLL_INTERVAL %v, must be at least 1 minute", data.PollInterval)
	}

	log.Printf("Using account %s, queryting %s to generate metric %s", data.AccountId, data.MetricName, data.NewMetricName)
	log.Printf("Poll interval is %s", data.PollInterval)

	// Create GraphQl client
	data.makeClient()

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sigs := <-sigs
		log.Printf("Process %v - Shutting down\n", sigs)
		os.Exit(0)
	}()

	data.queryGraphQl()

	data.makeMetrics()

	return

	// Start poll loop
	log.Println("Starting polling loop")
	for {
		startTime := time.Now()
		data.SampleTime = startTime.Unix()

		// Query timeslice metrics
		data.queryGraphQl()

		// Make results into metrics
		data.makeMetrics()

		remainder := data.PollInterval - time.Now().Sub(startTime)
		if remainder > 0 {
			log.Printf("Sleeping %v", remainder)

			// Wait remainder of poll interval
			time.Sleep(remainder)
		}
	}
}
