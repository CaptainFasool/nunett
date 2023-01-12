package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

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
	adapterMessage := models.AdapterMessage{}
	if err := json.Unmarshal([]byte(messageValue), &adapterMessage); err != nil {
		zlog.Error(err.Error())
	}

	switch adapterMessage.Data.Type {
	case "DeploymentRequest":
		zlog.Info("deployment request received")

		// send the depReq to DepReqQueue
		DepReqQueue <- adapterMessage

	case "DeploymentResponse":
		zlog.Info("deployment response received")

		// send the depRes to DepResQueue
		DepResQueue <- adapterMessage
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
		zlog.Error(fmt.Sprintf("Problem on dial: %s", err.Error()))
	}
	defer conn.Close()

	client := NewNunetAdapterClient(conn)
	emptyarg := ReceivedMessageParams{}
	msgStream, err := client.IncomingMessage(context.Background(), &emptyarg)

	if err != nil {
		zlog.Error(fmt.Sprintf("Problem Creating MsgStream! - : %s", err.Error()))
	}

	for {
		msg, err := msgStream.Recv()
		if err == io.EOF {
			zlog.Error("EOF")
		} else if err != nil {
			zlog.Error(fmt.Sprintf("Connection Problem - : %s", err.Error()))
			conn, err = grpc.Dial(utils.AdapterGrpcURL,
				[]grpc.DialOption{
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithBlock(),
					grpc.WithKeepaliveParams(kap)}...)
			if err != nil {
				zlog.Error(fmt.Sprintf("Problem on dial: %s", err.Error()))
			}
			client = NewNunetAdapterClient(conn)
			msgStream, err = client.IncomingMessage(context.Background(), &emptyarg)
			if err != nil {
				zlog.Error(fmt.Sprintf("Problem Creating MsgStream on refresh. - : %s", err.Error()))
			}
			time.Sleep(5 * time.Second)
		} else {
			zlog.Error(fmt.Sprintf("Received Message: : %s", msg))
			messageHandler(msg.MessageResponse)
		}
	}
}
