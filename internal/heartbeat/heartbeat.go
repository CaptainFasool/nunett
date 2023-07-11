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

	es, err := getElasticsearchClient()
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
		return
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
	es, err := getElasticsearchClient()
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
		return
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
	randomNumber := rand.Intn(100000) + 1
	docMap["callid"] = callid
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
		return
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		zlog.Sugar().Errorf("Error response received: %s", res.Status())
	}

}

func ProcessStatus(callid int, peerIDOfServiceHost string, serviceID string, status string, timestamp int) {
	es, err := getElasticsearchClient()
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
		return
	}

	indexName := "apm-nunet-dms-heartbeat"
	documentData := `{
		"peerIDOfServiceHost": 0,
		"serviceID": 0,
		"status": 0,
		"timestampCall": 0
		}`

	var docMap map[string]interface{}
	json.Unmarshal([]byte(documentData), &docMap)

	// Modify the timestamp field with the current timestamp
	docMap["timestamp"] = time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	docMap["peerIDOfServiceHost"] = peerIDOfServiceHost
	docMap["serviceID"] = serviceID
	docMap["status"] = status
	docMap["timestampCall"] = timestamp
	docMap["callid"] = callid

	// Extract the document ID from the response
	documentID, _ := getDocumentID("callid", callid)
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
		return
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		zlog.Sugar().Errorf("Error response received: %s", res.Status())
	}

}

func NtxPayment(callid int, serviceID string, successFailStatus string, peerID string, amountOfNtx int, timestamp int) {
	es, err := getElasticsearchClient()
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
		return
	}

	indexName := "apm-nunet-dms-heartbeat"
	documentData := `{
		"callid": 0,
		"serviceID": "",
		"successFailStatus": "",
		"peerID": "",
		"amountOfNtx":0,
		"timestampCall":""
		}`

	var docMap map[string]interface{}
	json.Unmarshal([]byte(documentData), &docMap)

	// Modify the timestamp field with the current timestamp
	docMap["timestamp"] = time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	docMap["peerID"] = peerID
	docMap["serviceID"] = serviceID
	docMap["successFailStatus"] = successFailStatus
	docMap["timestampCall"] = timestamp
	docMap["amountOfNtx"] = amountOfNtx
	docMap["callid"] = callid

	// Set a seed value based on the current time

	// Generate a random integer between 1 and 100
	documentID, _ := getDocumentID("callid", callid)

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
		return
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		zlog.Sugar().Errorf("Error response received: %s", res.Status())
	}

}

func DeviceResourceChange(cpu int, ram int) {
	es, err := getElasticsearchClient()
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
		return
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

	// Modify the timestamp field with the current timestamp
	docMap["timestamp"] = time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	docMap["cpu"] = cpu
	docMap["ram"] = ram
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

func getElasticsearchClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://dev.nunet.io:21001"}, // Elasticsearch server addresses
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getDocumentID(fieldName string, fieldValue int) (string, error) {
	// Connect to Elasticsearch
	es, err := getElasticsearchClient()
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
		return "", err
	}

	// Create a SearchRequest
	var body strings.Builder
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				fieldName: fieldValue,
			},
		},
	}
	if err := json.NewEncoder(&body).Encode(query); err != nil {
		return "", err
	}

	// Perform the search request
	res, err := esapi.SearchRequest{
		Index: []string{"your_index_name"},
		Body:  strings.NewReader(body.String()),
	}.Do(context.Background(), es)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// Check if any documents matched the query
	if res.IsError() {
		return "", nil
	}

	// Extract the document ID
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}

	hits := response["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) > 0 {
		documentID := hits[0].(map[string]interface{})["_id"].(string)
		return documentID, nil
	}

	return "", nil // Document not found
}
