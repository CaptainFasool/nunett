package heartbeat

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	"gitlab.com/nunet/device-management-service/libp2p"

	"math/rand"
	"strconv"

	"gitlab.com/nunet/device-management-service/utils"
)

var Done chan bool

func Heartbeat() {
	// Create a ticker that ticks every 1 minutes
	ticker := time.NewTicker(1 * time.Minute)

	// Start a goroutine to perform the repeated function calls
	go func() {
		for {
			select {
			case <-Done:
				// Stop the goroutine when the channel receives a signal
				return
			case <-ticker.C:
				defer func() {
					if r := recover(); r != nil {
						zlog.Sugar().Errorf("Recovered from error: %v", r)
					}
				}()
				Create()
			}
		}
	}()
}

func Create() {
	if libp2p.GetP2P().Host == nil {
		return
	}
	cfg := elasticsearch.Config{
		Addresses: []string{"http://dev.nunet.io:21001"},
		Username:  "admin",
		Password:  "changeme",
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
	}

	indexName := "apm-nunet-dms-heartbeat"

	documentID := libp2p.GetP2P().Host.ID().String()
	documentData := `{
		"cpu": 0,
		"ram": 0,
		"network": 0,
		"time": 0,
		"ID": "",
		"timestamp":""
		}`

	var docMap map[string]interface{}
	json.Unmarshal([]byte(documentData), &docMap)
	// get capacity user want to rent to NuNet

	metadata, _ := utils.ReadMetadataFile()

	// Modify the timestamp field with the current timestamp
	docMap["timestamp"] = time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	docMap["cpu"] = metadata.Reserved.CPU
	docMap["ram"] = metadata.Reserved.Memory
	docMap["ID"] = libp2p.GetP2P().Host.ID().String()

	updatedDocBytes, _ := json.Marshal(docMap)

	updatedDocString := string(updatedDocBytes)

	// Create the request
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(updatedDocString),
		Refresh:    "true",
	}

	// Perform the request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		zlog.Sugar().Errorf("Error indexing document: %v", err)
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		zlog.Sugar().Errorf("Error response received: %s", res.Status())
	}

}

func ProcessUsage(callid int, usedcpu int, usedram int, networkused int, timetaken int, ntx int) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://dev.nunet.io:21001"},
		Username:  "admin",
		Password:  "changeme",
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
	}

	indexName := "apm-nunet-dms-heartbeat"
	documentData := `{
		"usedcpu": 0,
		"usedram": 0,
		"usednetwork": 0,
		"timetaken": 0,
		"ID": "",
		"callid":0,
		"ntx":0,
		"timestamp":""
		}`

	var docMap map[string]interface{}
	json.Unmarshal([]byte(documentData), &docMap)

	// Modify the timestamp field with the current timestamp
	docMap["timestamp"] = time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	docMap["usedcpu"] = usedcpu
	docMap["usedram"] = usedram
	docMap["usednetwork"] = networkused
	docMap["timetaken"] = timetaken
	docMap["ntx"] = ntx

	// Set a seed value based on the current time

	// Generate a random integer between 1 and 100
	randomNumber := rand.Intn(100) + 1
	docMap["callid"] = randomNumber
	documentID := strconv.Itoa(randomNumber)

	docMap["ID"] = libp2p.GetP2P().Host.ID().String()

	updatedDocBytes, _ := json.Marshal(docMap)

	updatedDocString := string(updatedDocBytes)

	// Create the request
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(updatedDocString),
		Refresh:    "true",
	}

	// Perform the request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		zlog.Sugar().Errorf("Error indexing document: %v", err)
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		zlog.Sugar().Errorf("Error response received: %s", res.Status())
	}

}
