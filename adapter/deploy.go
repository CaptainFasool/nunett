package adapter

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"gitlab.com/nunet/device-management-service/docker"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// isMessageExpired check if message is old enough to be not processed
func isMessageExpired(message models.DeploymentRequest) bool {
	return time.Now().Unix()-message.Timestamp.Unix() > 5000
}

// RequestHandler is indirectly called by the polling function to process
// any new deployment request.
func RequestHandler(depReq models.DeploymentRequest, depRes models.DeploymentResponse) models.DeploymentResponse {
	// check for the service type and call the appropriate handler
	switch depReq.ServiceType {
	case "ml-training-gpu":
		log.Println("[deployment]: GPU deployemnt initiating")
		depRes := docker.HandleDeployment(depReq, depRes)
		return depRes
	case "cardano_node":
		// TO DO: call cardano_node trigger handler
		log.Println("[deployment]: Cardano Deployment initiating")
		return depRes
	default:
		log.Println("[deployment]: Deployment failed: Invalid ServiceType")
		return depRes
	}

}

func messageResponseHandler(message string) {
	// TODO: The following hack should not happen
	msg := `` + message + ``
	var messageArray []interface{}
	msgDoublQuote := strings.Replace(msg, `"`, "`", -1)               // replace double quote with backtick
	msgSingleQuote := strings.Replace(msgDoublQuote, `'`, `"`, -1)    // replace single quote with double quote
	formattedMessage := strings.Replace(msgSingleQuote, "`", `'`, -1) // replace backtick with single quote

	err := json.Unmarshal([]byte(formattedMessage), &messageArray)

	if err != nil {
		log.Fatal(err)
	}

	var deploymentRequest models.DeploymentRequest
	var deploymentResponse models.DeploymentResponse
	for _, message := range messageArray {
		messageObj := message.(map[string]interface{})
		messageValue := strings.Replace(messageObj["message"].(string), "'", `"`, -1) // replacing single quote to double quote to unmarshall
		json.Unmarshal([]byte(messageValue), &deploymentRequest)

		// if message is older, don't process it
		if isMessageExpired(deploymentRequest) {
			continue
		}
		depRes := RequestHandler(deploymentRequest, deploymentResponse)

		// If RequestHandler concludes that request for deployement has been accepted send a ack message to requester peer
		// Also in case of error send a message to the peer about failure of request
		NunetAdapterBaseClient("SendMessage", "MessageResponse", depRes)
	}
}

func PullMessages() {
	resp := NunetAdapterBaseClient("ReceivedMessage", "MessageResponse", &models.DeploymentRequest{})
	messageResponseHandler(resp.(MessageResponse).MessageResponse)
}

func NunetAdapterBaseClient(request string, response string, data interface{}) interface{} {
	conn, _ := grpc.Dial(utils.ADAPTER_GRPC_URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client := NewNunetAdapterClient(conn)

	// Call the adapter based on request and response params
	if request == "ReceivedMessage" && response == "MessageResponse" {
		res, _ := client.ReceivedMessage(ctx, &ReceivedMessageParams{})
		return res.MessageResponse
	} else if request == "SendMessage" && response == "MessageResponse" {
		deployementResponse := data.(models.DeploymentResponse)
		client.SendMessage(ctx, &MessageParams{
			NodeId: deployementResponse.NodeId,
			MessageContent: `{"success":"` + strconv.FormatBool(deployementResponse.Success) +
				`", "message":"` + deployementResponse.Content + `"}`,
		})
	}
	return struct{}{}
}

func PollAdapter() {
	for {
		// Run the function to retrieve messages from adapter in every 10 seconds
		time.Sleep(time.Second * 10)
		PullMessages()
	}
}
