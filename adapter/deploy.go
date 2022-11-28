package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/docker"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func messageHandler(message string) {
	messageValue := strings.Replace(message, `'`, `"`, -1)      // replacing single quote to double quote to unmarshall
	messageValue = strings.Replace(messageValue, `"{`, `{`, -1) // necessary because json becomes string in route and contains "key" : "{\"key\" : \"value\"}"
	messageValue = strings.Replace(messageValue, `}"`, `}`, -1) // necessary because json becomes string in route and contains "key" : "{\"key\" : \"value\"}"

	var adapterMessage models.AdapterMessage

	err := json.Unmarshal([]byte(messageValue), &adapterMessage)
	if err != nil {
		fmt.Println("            ", err)
		log.Println("Unable to parse received message: " + err.Error())
	}

	if adapterMessage.Message.ServiceType == "cardano_node" {
		// dowload kernel and filesystem files place them somewhere
		// TODO : organize fc files
		utils.DownloadFile("https://d.nunet.io/fc/vmlinux", "/etc/nunet/vmlinux")
		utils.DownloadFile("https://d.nunet.io/fc/nunet-fc-ubuntu-20.04-0.ext4", "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4")

		// makerequest to start-default with downloaded files.
		type StartDefaultBody struct {
			KernelImagePath string `json:"kernel_image_path"`
			FilesystemPath  string `json:"filesystem_path"`
		}
		startDefaultBody := &StartDefaultBody{
			KernelImagePath: "/etc/nunet/vmlinux",
			FilesystemPath:  "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4",
		}
		jsonBody, _ := json.Marshal(startDefaultBody)

		resp := utils.MakeInternalRequest(&gin.Context{}, "POST", "/vm/start-default", jsonBody)
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}
		log.Println(string(respBody))

		selfNodeId, _ := getSelfNodeID()
		depResp := models.DeploymentResponse{
			NodeId:  selfNodeId,
			Success: true,
			Content: "Cardano Node Deployed",
		}
		jsonDepResp, _ := json.Marshal(depResp)
		SendMessage(adapterMessage.Sender, string(jsonDepResp))

	} else if adapterMessage.Message.ServiceType == "ml-training-gpu" {
		depResp := models.DeploymentResponse{}
		depResp = docker.HandleDeployment(adapterMessage.Message, depResp)

		jsonDepResp, _ := json.Marshal(depResp)
		SendMessage(adapterMessage.Sender, string(jsonDepResp))
	} else {
		log.Println("Error: Uknown service type - ", adapterMessage.Message.ServiceType)
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
