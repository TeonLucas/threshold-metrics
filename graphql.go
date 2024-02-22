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
	timestamp := int64(result["timestamp"].(float64))
	metricValue := result[data.MetricName]
	facetMap := make(map[string]string)
	facetMap["entity.guid"] = fmt.Sprintf("%v", result["entity.guid"])
	facets := strings.Split(data.MetricFacet, ",")
	for _, facet := range facets {
		key := strings.TrimSpace(facet)
		facetMap[key] = fmt.Sprintf("%v", result[key])
	}
	log.Printf("DEBUG facets %v: timestamp=%d value=%v", facetMap, timestamp, metricValue)
}

func (data *AccountData) queryGraphQl() {
	var err error
	var gQuery GraphQlPayload
	var j []byte

	// Make graphQl request to lookup entity names by guid (if not already cached)
	query := fmt.Sprintf("SELECT %s,%s,entity.guid FROM Metric WHERE %s LIMIT MAX SINCE 10 minutes ago",
		data.MetricName, data.MetricFacet, data.MetricWhere)
	log.Printf("DEBUG NRQL query: %q", query)
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
	for i, result := range graphQlResult.Data.Actor.Account.Nrql.Results {
		log.Printf("DEBUG parsing result %d", i+1)
		data.parseResult(result)
	}
}
