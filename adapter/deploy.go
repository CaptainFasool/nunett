package adapter

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func messageHandler(message string) {
	messageValue := strings.Replace(message, `'`, `"`, -1)      // replacing single quote to double quote to unmarshall
	messageValue = strings.Replace(messageValue, `"{`, `{`, -1) // necessary because json becomes string in route and contains "key" : "{\"key\" : \"value\"}"
	messageValue = strings.Replace(messageValue, `}"`, `}`, -1) // necessary because json becomes string in route and contains "key" : "{\"key\" : \"value\"}"

	// interpret the kind of message and push it to specific queue
	var objmap map[string]interface{}

	if err := json.Unmarshal([]byte(messageValue), &objmap); err != nil {
		zlog.Error(err.Error())
	}

	switch objmap["message_type"] {
	// TODO: avoid hardcoding of cases here. Better have a map of message type and model/struct
	// to unmarshal on
	// TODO: For each message_type, unmarshal `message` and send it to specific channel
	case "DeploymentRequest":
		zlog.Info("deployment request received")

		var depReq models.DeploymentRequest
		mapstructure.Decode(objmap["message"], &depReq)

		// send the depReq to DepReqQueue
		DepReqQueue <- depReq

	case "DeploymentResponse":
		zlog.Info("deployment response received")

		var depRes models.DeploymentResponse
		mapstructure.Decode(objmap["message"], &depRes)

		// send the depRes to depReq queue
		DepResQueue <- depRes

	}

}

func StartMessageReceiver() {
	kap := keepalive.ClientParameters{
		Time:                time.Duration(15 * time.Minute), // ping server every 15 minutes
		Timeout:             time.Duration(60 * time.Second), // close connection after 60 seconds of no-ackowledgement for ping
		PermitWithoutStream: true,                            // allow keepalive ping even without data
	}
	conn, err := grpc.Dial(utils.AdapterGrpcURL, []grpc.DialOption{
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
			conn, err = grpc.Dial(utils.AdapterGrpcURL,
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
