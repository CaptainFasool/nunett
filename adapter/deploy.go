package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	// "gitlab.com/nunet/device-management-service/gpu"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func messageHandler(message string) {
	/*
		Preliminary formatting of messages for deployment requests and othre functions.
		Data structure and specific flags need to be standardized.

			message contains four sections.
				- sender: nodeid of sender
				- mess/ge: the actual message sent
				- uid: uid of the message
				- msg_time: date time in iso format

			deployment request indicated by the prefix: "depreq"
			options for deployment request:
				- cardano_node
				- machine_learning

			acknowledgement of deployment request indicated by the prefix: "ackdepreq"
			followed by the type of the request then followed by "success" or "fail"

		TODO:
			- error handling
			- better data structure for functional messages such as deployment requests and results
	*/

	messageValue := strings.Replace(message, "'", `"`, -1) // replacing single quote to double quote to unmarshall
	var adapterMessage models.AdapterMessage
	json.Unmarshal([]byte(messageValue), &adapterMessage)

	// remove after testing
	fmt.Println("Parsed Message: ")
	fmt.Println("Sender: ", adapterMessage.Sender)
	fmt.Println("Time: ", adapterMessage.Timestamp)
	fmt.Println("UID: ", adapterMessage.Uid)
	fmt.Println("Message: ", adapterMessage.Message)
	// remove after testing

	if strings.HasPrefix(adapterMessage.Message, "depreq") { // deployment request
		if strings.HasSuffix(adapterMessage.Message, "cardano_node") { // cardano node
			// attempt to run vm and reply success or failure
		} else if strings.HasSuffix(adapterMessage.Message, "machine_learning") { // machine learning
			// attempt to run a container and reply success or failure
		} else {
			SendMessage(adapterMessage.Sender, "ackdepreq unknown")
		}
	}
}

func StartMessageReceiver() {
	kap := keepalive.ClientParameters{
		Time:                time.Duration(15 * time.Minute), // ping server every 15 minutes
		Timeout:             time.Duration(60 * time.Second), // close connection after 60 seconds of no-ackowledgement for ping
		PermitWithoutStream: true,                            // allow keepalive ping even without data
	}
	conn, err := grpc.Dial(utils.ADAPTER_GRPC_URL, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(kap)}...)
	if err != nil {
		log.Println("[ADAPTER MSG] ERROR: Problem on dial", err.Error())
	}
	defer conn.Close()

	client := NewNunetAdapterClient(conn)
	emptyarg := ReceivedMessageParams{}
	msgStream, err := client.IncomingMessage(context.Background(), &emptyarg)

	if err != nil {
		log.Println("[ADAPTER MSG] ERROR: Problem Creating MsgStream! - ", err.Error())
	}

	for {
		msg, err := msgStream.Recv()
		if err == io.EOF {
			log.Println("[ADAPTER MSG] ERROR: EOF")
		} else if err != nil {
			log.Println("[ADAPTER MSG] ERROR: Connection Problem - ", err.Error())
			conn, err = grpc.Dial(utils.ADAPTER_GRPC_URL,
				[]grpc.DialOption{
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithBlock(),
					grpc.WithKeepaliveParams(kap)}...)
			if err != nil {
				log.Println("[ADAPTER MSG] ERROR: Problem on dial: ", err.Error())
			}
			client = NewNunetAdapterClient(conn)
			msgStream, err = client.IncomingMessage(context.Background(), &emptyarg)
			if err != nil {
				log.Println("[ADAPTER MSG] ERROR: Problem Creating MsgStream on refresh. - ", err.Error())
			}
			time.Sleep(5 * time.Second)
		} else {
			log.Println("[ADAPTER MSG] INFO: Received Message: ", msg)
			messageHandler(msg.MessageResponse)
		}
	}
}
