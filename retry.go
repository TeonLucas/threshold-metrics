package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Make API request with error retry
func retryQuery(client *http.Client, method, url, data string, headers []string) (b []byte) {
	var res *http.Response
	var err error
	var body io.Reader

	if len(data) > 0 {
		body = strings.NewReader(data)
	}

	// up to 3 retries on API error
	for j := 1; j <= 3; j++ {
		req, _ := http.NewRequest(method, url, body)
		for _, h := range headers {
			params := strings.Split(h, ":")
			req.Header.Set(params[0], params[1])
		}
		res, err = client.Do(req)
		if err != nil {
			log.Println(err)
		}
		if res != nil {
			if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusAccepted {
				break
			}
			log.Printf("Retry %d: http status %d", j, res.StatusCode)
		} else {
			log.Printf("Retry %d: no response", j)
		}
		time.Sleep(500 * time.Millisecond)
	}
	b, err = io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	res.Body.Close()
	return
}
