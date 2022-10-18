package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	// "gitlab.com/nunet/device-management-service/gpu"
	"gitlab.com/nunet/device-management-service/models"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func verifyMessage(message models.DeploymentRequest) error {

	// check message and return error if we need not to process it.
	if time.Now().Unix()-message.Timestamp.Unix() > 5000 {
		fmt.Println("----------------------Expired Message------------------------------")
		return errors.New("message expired")
	}
	return nil
}

func requestHandler(message models.DeploymentRequest) error {

	err := verifyMessage(message)

	if err != nil {
		return err
	}

	// check for the service type and call the appropriate handler
	switch message.ServiceType {
	case "ml-training-gpu":
		{
			// TO DO: call GPU trigger handler
			fmt.Println("++++++++++++++++++++++++++GPU deployemnt+++++++++++++++++")
			return nil
		}
	case "cardano_node":
		{
			// TO DO: call cardano_node trigger handler
			fmt.Println("++++++++++++++++++++++++++Cardano Deployment+++++++++++++++++")
			return nil
		}
	default:
		{
			fmt.Println("----------------------Invalid service Type------------------------------")
			return errors.New("invalid service type")
		}
	}

}

func messageResponseHandler(message string) {
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
	for _, message := range messageArray {
		messageObj := message.(map[string]interface{})
		messageValue := strings.Replace(messageObj["message"].(string), "'", `"`, -1) // replacing single quote to double quote to unmarshall
		json.Unmarshal([]byte(messageValue), &deploymentRequest)
		err := requestHandler(deploymentRequest)

		// If requestHandler concludes that request for deployement has been accepted send a ack message to requester peer
		// Also in case of error send a message to the peer about failure of request
		if err != nil {
			content := fmt.Sprint(err)
			NunetAdapterBaseClient("SendMessage", "MessageResponse", &models.DeploymentResponse{
				NodeId:  deploymentRequest.AddressUser,
				Success: false,
				Content: content,
			})
		} else {
			NunetAdapterBaseClient("SendMessage", "MessageResponse", &models.DeploymentResponse{
				NodeId:  deploymentRequest.AddressUser,
				Success: true,
			})
		}
	}
}

func PullMessages() {
	resp := NunetAdapterBaseClient("ReceivedMessage", "MessageResponse", &models.DeploymentRequest{})
	messageResponseHandler(resp.(MessageResponse).MessageResponse)
}

func NunetAdapterBaseClient(request string, response string, data interface{}) interface{} {
	adapter_address := "localhost:60777"
	conn, _ := grpc.Dial(adapter_address, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
