package heartbeat

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/elastic/go-elasticsearch/v8"
	"gitlab.com/nunet/device-management-service/db"

	"strconv"

	"gitlab.com/nunet/device-management-service/models"

	"gitlab.com/nunet/device-management-service/utils"
)

var Done chan bool

func Heartbeat() {
	// Create a ticker that ticks every 1 minutes
	ticker := time.NewTicker(1 * time.Minute)
	// Start a goroutine to perform the repeated function calls
	go func() {
		defer func() {
			if r := recover(); r != nil {
				zlog.Sugar().Errorf("Recovered from error: %v", r)
			}
		}()
		for {
			select {
			case <-Done:
				// Stop the goroutine when the channel receives a signal
				return
			case <-ticker.C:
				err := Create()
				if err != nil {
					zlog.Sugar().Errorln(err)
				}
			}
		}
	}()
}

func Create() error {
	if !utils.ReadyForElastic() {
		zlog.Warn("Elastic search is not ready yet")
		return nil
	}
	indexName := "apm-nunet-dms-heartbeat"

	documentID := elastictoken.NodeId
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
	docMap["ID"] = elastictoken.NodeId

	updatedDocBytes, _ := json.Marshal(docMap)

	updatedDocString := string(updatedDocBytes)

	// Create the request
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(updatedDocString),
		Refresh:    "true",
	}

	if esClient == nil {
		client, err := getElasticSearchClient()
		if err != nil {
			return fmt.Errorf("unable to create esClient token: %v", err)
		}
		esClient = client
	}

	// Perform the request
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		return fmt.Errorf("Error indexing document: %v", err)
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		return fmt.Errorf("Error response received: %s", res.Status())
	}
	return nil
}

func ProcessUsage(callid int, usedcpu int, usedram int, networkused int, timetaken int, ntx int) error {
	if !utils.ReadyForElastic() {
		return errors.New("Elasticsearch is not ready")
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

	if esClient == nil {
		client, err := getElasticSearchClient()
		if err != nil {
			return fmt.Errorf("unable to create esClient token: %v", err)
		}
		esClient = client
	}

	exists, err := documentExists(context.Background(), esClient, indexName, documentID)

	if err != nil {
		return fmt.Errorf("Error retrieving the document : %v", err)
	}

	if exists {

		fields := map[string]interface{}{
			"callid":      callid,
			"usedcpu":     usedcpu,
			"usedram":     usedram,
			"usednetwork": networkused,
			"timetaken":   timetaken,
			"timestamp":   time.Now().Format("2006-01-02T15:04:05.999Z07:00"),

			"ntx": ntx,
		}

		//err = UpdateDocumentField(context.Background(), es, indexName, documentID, "usedcpu", "50")
		err = updateDocumentFields(context.Background(), esClient, indexName, documentID, fields)
		if err != nil {
			return fmt.Errorf("Error retrieving the document : %v", err)
		}
		return nil
	}

	docMap["ID"] = elastictoken.NodeId

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
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		return fmt.Errorf("Error indexing document: %v", err)
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		return fmt.Errorf("Error response received: %s", res.Status())
	}
	return nil
}

func ProcessStatus(callid int, peerIDOfServiceHost string, serviceID string, status string, timestamp int) error {
	if !utils.ReadyForElastic() {
		return errors.New("Elasticsearch is not ready")
	}
	indexName := "apm-nunet-dms-heartbeat"

	documentID := strconv.Itoa(callid)

	if esClient == nil {
		client, err := getElasticSearchClient()
		if err != nil {
			return fmt.Errorf("unable to create esClient token: %v", err)
		}
		esClient = client
	}

	exists, err := documentExists(context.Background(), esClient, indexName, documentID)

	if err != nil {
		return fmt.Errorf("Error retrieving the document : %v", err)
	}

	if exists {

		fields := map[string]interface{}{
			"timestamp": time.Now().Format("2006-01-02T15:04:05.999Z07:00"),
			"callid":    callid,
			"status":    status,
			"serviceID": serviceID,
		}
		err = updateDocumentFields(context.Background(), esClient, indexName, documentID, fields)

		if err != nil {
			return fmt.Errorf("Error updating the document : %v", err)
		}
		return nil
	}
	return nil
}

func NtxPayment(callid int, serviceID string, successFailStatus string, peerID string, amountOfNtx int, timestamp int) error {
	if !utils.ReadyForElastic() {
		return errors.New("Elasticsearch is not ready")
	}
	indexName := "apm-nunet-dms-heartbeat"

	documentID := strconv.Itoa(callid)

	if esClient == nil {
		client, err := getElasticSearchClient()
		if err != nil {
			return fmt.Errorf("unable to create esClient token: %v", err)
		}
		esClient = client
	}

	exists, err := documentExists(context.Background(), esClient, indexName, documentID)

	if err != nil {
		return fmt.Errorf("Error retrieving the document : %v", err)
	}

	if exists {

		fields := map[string]interface{}{
			"callid":    callid,
			"serviceID": serviceID,
			"status":    successFailStatus,
			"ntx":       amountOfNtx,
		}
		err = updateDocumentFields(context.Background(), esClient, indexName, documentID, fields)

		if err != nil {
			return fmt.Errorf("Error updating the document : %v", err)
		}
		return nil
	}
	return nil
}

func DeviceResourceChange(cpu int, ram int) error {
	if !utils.ReadyForElastic() {
		return errors.New("Elasticsearch is not ready")
	}
	indexName := "apm-nunet-dms-heartbeat"

	documentID := elastictoken.NodeId

	if esClient == nil {
		client, err := getElasticSearchClient()
		if err != nil {
			return fmt.Errorf("unable to create esClient token: %v", err)
		}
		esClient = client
	}

	exists, err := documentExists(context.Background(), esClient, indexName, documentID)

	if err != nil {
		return fmt.Errorf("Error retrieving the document : %v", err)
	}

	if exists {

		fields := map[string]interface{}{
			"cpu": cpu,
			"ram": ram,
		}
		err = updateDocumentFields(context.Background(), esClient, indexName, documentID, fields)

		if err != nil {
			return fmt.Errorf("Error updating the document : %v", err)
		}
		return nil
	}
	return nil
}

func DmsLoggs(log string, level string) error {
	if !utils.ReadyForElastic() {
		return errors.New("Elasticsearch is not ready")
	}
	indexName := "apm-nunet-dms-logs"
	documentID := strconv.Itoa(generateRandomInt())
	documentData := `{
		"log": "",
		"level": "",
		"ID": "",
		"timestamp":""
		}`

	var docMap map[string]interface{}
	json.Unmarshal([]byte(documentData), &docMap)

	docMap["timestamp"] = time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	docMap["log"] = log
	docMap["level"] = level
	docMap["ID"] = elastictoken.NodeId
	updatedDocBytes, _ := json.Marshal(docMap)

	updatedDocString := string(updatedDocBytes)

	// Create the request
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(updatedDocString),
		Refresh:    "true",
	}

	if esClient == nil {
		client, err := getElasticSearchClient()
		if err != nil {
			return fmt.Errorf("unable to create esClient token: %v", err)
		}
		esClient = client
	}

	// Perform the request
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		return fmt.Errorf("Error indexing document: %v", err)
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		return fmt.Errorf("Error response received: %s", res.Status())
	}
	return nil
}

func getElasticSearchClient() (*elasticsearch.Client, error) {
	elastictoken = models.ElasticToken{}
	db.DB.Where("channel_name = ?", utils.GetChannelName()).Find(&elastictoken)
	accessToken := elastictoken.Token

	if accessToken == "" {
		accessToken, err := NewToken(elastictoken.NodeId, elastictoken.ChannelName)
		if err != nil {
			return nil, fmt.Errorf("unable to create token: %v", err)
		}
		elastictoken.Token = accessToken
	}

	var address string
	if elastictoken.ChannelName == "nunet-test" {
		address = "https://elastic-test.test.nunet.io"
	} else {
		address = "http://elastic-staging.dev.nunet.io"
	}

	if esClient != nil {
		return esClient, nil
	}

	// Create the HTTP client with the access token
	var err error
	esClient, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
		Header:    http.Header{"Authorization": []string{"ApiKey " + accessToken}},
	})
	if err != nil {
		return nil, err
	}
	// Check if the connection is successful
	res, err := esClient.Info()
	if err != nil {
		return nil, err
	}
	// defer res.Body.Close()

	// Decode the response
	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, err
	}

	return esClient, nil
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

func GetToken(peerID string, channel string) (string, error) {
	data := url.Values{}
	data.Set("peerid", peerID)
	url := ""
	if channel == "nunet-test" {
		url = "https://elastic-test.test.nunet.io/getOrCreate"
	} else {
		url = "http://dev.nunet.io:24000/getOrCreate"
	}
	resp, err := http.PostForm(url, data)
	if err != nil {
		return "", fmt.Errorf("failed to make a request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading respones: %v", err)
	}

	return string(body), nil
}

func NewToken(peerID string, channel string) (string, error) {
	zlog.Sugar().Infof("obtaining a new elastic token")
	var existingTokens []models.ElasticToken
	result := db.DB.Where("channel_name = ?", channel).Find(&existingTokens)
	if result.RowsAffected > 0 {
		db.DB.Delete(&existingTokens)
	}

	if peerID == "" || channel == "" {
		return "", fmt.Errorf("peerID and channel can't be empty")
	}

	token, _ := GetToken(peerID, utils.GetChannelName())
	newElastictoken := models.ElasticToken{NodeId: peerID, ChannelName: channel, Token: token}
	result = db.DB.Create(&newElastictoken)
	if result.Error != nil {
		zlog.Sugar().Errorf("could not create elastic search access token record in DB: %v", result.Error)
		return "", result.Error
	}
	elastictoken = newElastictoken
	return token, nil
}

func CheckToken(peerID string, channel string) error {
	elasticToken := models.ElasticToken{ChannelName: channel}
	result := db.DB.Where("channel_name = ?", channel).Find(&elasticToken)
	if result.Error != nil || elasticToken.Token == "" {
		zlog.Sugar().Warnf("unable to read elastic token from db - err: %v , noToken: %t", result.Error, elasticToken.Token == "")
		_, err := NewToken(peerID, channel)
		if err != nil {
			return fmt.Errorf("unable to create token: %v", err)
		}
	}

	return nil
}

func generateRandomInt() int {
	min := big.NewInt(1000000000)
	max := big.NewInt(10000000000 - 1)
	n, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		zlog.Sugar().Errorf("unable to generate random integer: %v", err)
	}
	return int(n.Int64()) + 1000000000
}
