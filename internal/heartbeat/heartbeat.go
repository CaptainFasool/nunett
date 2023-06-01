package heartbeat

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"strings"
	"time"
)

func Heartbeat() {
	// Create a ticker that ticks every 2 minutes
	ticker := time.NewTicker(1 * time.Minute)

	// Create a channel to receive ticks from the ticker
	done := make(chan bool)
	// Start a goroutine to perform the repeated function calls
	go func() {
		for {
			select {
			case <-done:
				// Stop the goroutine when the channel receives a signal
				return
			case <-ticker.C:
				// Call your function here
				Create()
			}
		}
	}()
}

func Create() {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://dev.nunet.io:21001"},
		Username:  "",
		Password:  "",
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the Elasticsearch client: %s", err)
	}

	indexName := "nunet-dms"
	documentID := "1"
	documentData := `{
		"cpu": "60",
		"ram": "120",
		"network": "180",
		"time": "240",
		"ID": "unique-id"

		}`

	// Create the request
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(documentData),
		Refresh:    "true",
	}

	// Perform the request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error indexing document: %s", err)
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		log.Fatalf("Error response received: %s", res.Status())
	}

}
