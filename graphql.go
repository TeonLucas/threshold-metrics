package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

const (
	GraphQlEndpoint = "https://api.newrelic.com/graphql"
	GraphQlQuery    = "{actor {account(id: %s) {nrql (query: %q) {results}}}}"
)

type GraphQlPayload struct {
	Query string `json:"query"`
}

type GraphQlResult struct {
	Data struct {
		Actor struct {
			Account struct {
				Nrql struct {
					Results []NrqlResult `json:"results"`
				} `json:"nrql"`
			} `json:"account"`
		} `json:"actor"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type NrqlResult map[string]interface{}

func (data *AccountData) parseResult(result NrqlResult) {
	var ok, valid bool
	var value, timestampRaw, timeslice interface{}

	// Get timestamp
	timestampRaw, ok = result["timestamp"]
	if !ok {
		return
	}
	timestamp := int64(timestampRaw.(float64))

	// Get metric aggretate values
	timeslice, ok = result[data.MetricName]
	if !ok {
		return
	}

	// Get attributes
	attributes := make(map[string]string)
	for _, key := range data.Attributes {
		value, ok = result[key]
		if ok {
			attributes[key] = fmt.Sprintf("%v", value)
		}
		if ok && key == "entity.guid" {
			valid = true
		}
	}

	// Make sure there is a Guid
	if valid {
		data.pushMetric(timestamp, timeslice, attributes)
	}
	// Advance the timestamp for next query
	if timestamp > data.Timestamp {
		data.Timestamp = timestamp
	}
}

func (data *AccountData) queryGraphQl() {
	var err error
	var gQuery GraphQlPayload
	var j []byte

	// Make graphQl request to lookup entity names by guid (if not already cached)
	query := fmt.Sprintf("SELECT %s, %s FROM Metric WHERE %s AND timestamp > %d LIMIT MAX SINCE %d minutes ago",
		data.MetricName, strings.Join(data.Attributes, ", "), data.MetricWhere, data.Timestamp, data.Since)

	//log.Printf("DEBUG NRQL query: %q", query)

	gQuery.Query = fmt.Sprintf(GraphQlQuery, data.AccountId, query)
	j, err = json.Marshal(gQuery)
	if err != nil {
		log.Printf("Error creating GraphQl query: %v", err)
	}

	b := retryQuery(data.Client, "POST", GraphQlEndpoint, string(j), data.GraphQlHeaders)
	var graphQlResult GraphQlResult
	log.Printf("Parsing response %d bytes", len(b))
	err = json.Unmarshal(b, &graphQlResult)
	if err != nil {
		log.Printf("Error parsing GraphQl result: %v", err)
	}
	for _, result := range graphQlResult.Data.Actor.Account.Nrql.Results {
		data.parseResult(result)
	}
}
