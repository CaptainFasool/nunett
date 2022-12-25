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
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/docker"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/statsdb"
	"gitlab.com/nunet/device-management-service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func sendDeploymentResponse(address string, success bool, message string) {
	selfNodeId, err := getSelfNodeID()
	if err != nil {
		log.Println("Error: Unable to get self node id. ", err)
	}
	jsonDepResp, _ := json.Marshal(models.DeploymentResponse{
		NodeId:  selfNodeId,
		Success: success,
		Content: message,
	})
	SendMessage(address, string(jsonDepResp))
}

func messageHandler(message string) {
	messageValue := strings.Replace(message, `'`, `"`, -1)      // replacing single quote to double quote to unmarshall
	messageValue = strings.Replace(messageValue, `"{`, `{`, -1) // necessary because json becomes string in route and contains "key" : "{\"key\" : \"value\"}"
	messageValue = strings.Replace(messageValue, `}"`, `}`, -1) // necessary because json becomes string in route and contains "key" : "{\"key\" : \"value\"}"

	var adapterMessage models.AdapterMessage

	err := json.Unmarshal([]byte(messageValue), &adapterMessage)
	if err != nil {
		fmt.Println("            ", err)
		log.Println("Error: Unable to parse received message: " + err.Error())
	}

	if adapterMessage.Message.ServiceType == "cardano_node" {
		// dowload kernel and filesystem files place them somewhere
		// TODO : organize fc files
		kernelFileUrl := "https://d.nunet.io/fc/vmlinux"
		kernelFilePath := "/etc/nunet/vmlinux"
		filesystemUrl := "https://d.nunet.io/fc/nunet-fc-ubuntu-20.04-0.ext4"
		filesystemPath := "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4"
		pKey := adapterMessage.Message.Params.PublicKey
		nodeId := adapterMessage.Message.Params.NodeID

		err = utils.DownloadFile(kernelFileUrl, kernelFilePath)
		if err != nil {
			log.Println("Error: Downloading ", kernelFileUrl, " - ", err.Error())
			sendDeploymentResponse(adapterMessage.Sender,
				false,
				fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", kernelFileUrl))
		}
		err = utils.DownloadFile(filesystemUrl, filesystemPath)
		if err != nil {
			log.Println("Error: Downloading ", filesystemUrl, " - ", err.Error())
			sendDeploymentResponse(adapterMessage.Sender,
				false,
				fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", filesystemUrl))
		}

		// makerequest to start-default with downloaded files.
		startDefaultBody := struct {
			KernelImagePath string `json:"kernel_image_path"`
			FilesystemPath  string `json:"filesystem_path"`
			PublicKey       string `json:"public_key"`
			NodeID          string `json:"node_id"`
		}{
			KernelImagePath: kernelFilePath,
			FilesystemPath:  filesystemPath,
			PublicKey:       pKey,
			NodeID:          nodeId,
		}
		jsonBody, _ := json.Marshal(startDefaultBody)

		resp := utils.MakeInternalRequest(&gin.Context{}, "POST", "/vm/start-default", jsonBody)
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			sendDeploymentResponse(adapterMessage.Sender, false, "Cardano Node Deployment Failed. Unable to spawn VM.")
		}
		log.Println(string(respBody))

	} else if adapterMessage.Message.ServiceType == "ml-training-gpu" {
		depResp := models.DeploymentResponse{}

		callID := statsdb.GetCallID()
		peerIDOfServiceHost := adapterMessage.Message.Params.NodeID
		timeStamp := float32(statsdb.GetTimestamp())
		status := "accepted"

		ServiceCallParams := models.ServiceCall{
			CallID:              callID,
			PeerIDOfServiceHost: peerIDOfServiceHost,
			ServiceID:           adapterMessage.Message.ServiceType,
			CPUUsed:             float32(adapterMessage.Message.Constraints.CPU),
			MaxRAM:              float32(adapterMessage.Message.Constraints.Vram),
			MemoryUsed:          float32(adapterMessage.Message.Constraints.RAM),
			NetworkBwUsed:       0.0,
			TimeTaken:           0.0,
			Status:              status,
			Timestamp:           timeStamp,
		}
		statsdb.ServiceCall(ServiceCallParams)

		requestTracker := models.RequestTracker{
			ServiceType: adapterMessage.Message.ServiceType,
			CallID:      callID,
			NodeID:      peerIDOfServiceHost,
			Status:      status,
		}
		result := db.DB.Create(&requestTracker)
		if result.Error != nil {
			panic(result.Error)
		}

		depResp = docker.HandleDeployment(adapterMessage.Message, depResp)
		jsonDepResp, _ := json.Marshal(depResp)
		SendMessage(adapterMessage.Sender, string(jsonDepResp))
	} else {
		log.Println("Error: Uknown service type - ", adapterMessage.Message.ServiceType)
		sendDeploymentResponse(adapterMessage.Sender, false, "Unknown service type.")
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
