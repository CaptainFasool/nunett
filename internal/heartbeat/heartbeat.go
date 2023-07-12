package heartbeat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"gitlab.com/nunet/device-management-service/libp2p"

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
		"timestamp":"",
		"serviceID":""

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

	docMap["callid"] = callid
	documentID := strconv.Itoa(callid)

	exists, err := documentExists(context.Background(), es, indexName, documentID)

	if err != nil {
		zlog.Sugar().Errorf("Error retrieving the document : %v", err)
		return
	}

	if exists {

		fields := map[string]interface{}{
			"callid":      callid,
			"usedcpu":     usedcpu,
			"usedram":     usedram,
			"usednetwork": networkused,
			"timetaken":   timetaken,
			"ntx":         ntx,
		}
		//err = UpdateDocumentField(context.Background(), es, indexName, documentID, "usedcpu", "50")
		err = updateDocumentFields(context.Background(), es, indexName, documentID, fields)

		if err != nil {
			zlog.Sugar().Errorf("Error retrieving the document : %v", err)
		}
		return
	}

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

	documentID := strconv.Itoa(callid)

	exists, err := documentExists(context.Background(), es, indexName, documentID)

	if err != nil {
		zlog.Sugar().Errorf("Error retrieving the document : %v", err)
		return
	}

	if exists {

		fields := map[string]interface{}{
			"timestamp": time.Now().Format("2006-01-02T15:04:05.999Z07:00"),
			"callid":    callid,
			"status":    status,
			"serviceID": serviceID,
		}
		err = updateDocumentFields(context.Background(), es, indexName, documentID, fields)

		if err != nil {
			zlog.Sugar().Errorf("Error updating the document : %v", err)
		}
		return
	}

}

func NtxPayment(callid int, serviceID string, successFailStatus string, peerID string, amountOfNtx int, timestamp int) {
	es, err := getElasticsearchClient()
	if err != nil {
		zlog.Sugar().Errorf("Error creating the Elasticsearch client: %v", err)
		return
	}

	indexName := "apm-nunet-dms-heartbeat"

	documentID := strconv.Itoa(callid)

	exists, err := documentExists(context.Background(), es, indexName, documentID)

	if err != nil {
		zlog.Sugar().Errorf("Error retrieving the document : %v", err)
		return
	}

	if exists {

		fields := map[string]interface{}{
			"callid":    callid,
			"serviceID": serviceID,
			"status":    successFailStatus,
			"ntx":       amountOfNtx,
		}
		err = updateDocumentFields(context.Background(), es, indexName, documentID, fields)

		if err != nil {
			zlog.Sugar().Errorf("Error updating the document : %v", err)
		}
		return
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

	exists, err := documentExists(context.Background(), es, indexName, documentID)

	if err != nil {
		zlog.Sugar().Errorf("Error retrieving the document : %v", err)
		return
	}

	if exists {

		fields := map[string]interface{}{
			"cpu": cpu,
			"ram": ram,
		}
		err = updateDocumentFields(context.Background(), es, indexName, documentID, fields)

		if err != nil {
			zlog.Sugar().Errorf("Error updating the document : %v", err)
		}
		return
	}

}

func getElasticsearchClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://dev.nunet.io:21001"}, // Elasticsearch server addresses
		Username:  "admin",
		Password:  "changeme",
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// DocumentExists checks if a document ID exists in Elasticsearch
func documentExists(ctx context.Context, es *elasticsearch.Client, index, docID string) (bool, error) {
	req := esapi.ExistsRequest{
		Index:      index,
		DocumentID: docID,
	}

	res, err := req.Do(ctx, es)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return false, nil
	}
	return res.StatusCode == 200, nil
}

func toRequestBody(script string, params map[string]interface{}) *strings.Reader {
	body := map[string]interface{}{
		"script": map[string]interface{}{
			"source": script,
			"params": params,
		},
	}

	jsonBody, _ := json.Marshal(body)
	return strings.NewReader(string(jsonBody))
}

func updateDocumentFields(ctx context.Context, es *elasticsearch.Client, index, docID string, fields map[string]interface{}) error {
	// Prepare the update script
	script := ""
	params := make(map[string]interface{})
	for field, value := range fields {
		script += fmt.Sprintf("ctx._source.%s = params.%s;", field, field)
		params[field] = value
	}

	// Build the update request
	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: docID,
		Body:       toRequestBody(script, params),
	}

	// Perform the update request
	res, err := req.Do(ctx, es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Check the response
	if res.IsError() {
		return nil
	}

	return nil
}
