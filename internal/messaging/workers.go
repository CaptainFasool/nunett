// Package messaging contains logic to listen to queues/channels spread over the application.
package messaging

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/docker"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/statsdb"
	"gitlab.com/nunet/device-management-service/utils"
)

func sendDeploymentResponse(address string, success bool, message string) {
	jsonDepResp, _ := json.Marshal(models.DeploymentResponse{
		Success: success,
		Content: message,
	})
	adapter.SendMessage(address, string(jsonDepResp))
}

func DeploymentWorker() {
	for {
		select {
		case msg := <-adapter.DepReqQueue:
			var adapterMessage models.AdapterMessage
			var depReq models.DeploymentRequest

			mapstructure.Decode(msg, &adapterMessage)
			mapstructure.Decode(msg.Data.Message, &depReq)

			sender := adapterMessage.Sender

			if depReq.ServiceType == "cardano_node" {
				handleCardanoDeployment(depReq, sender)
			} else if depReq.ServiceType == "ml-training-gpu" {
				handleGpuDeployment(depReq, sender)
			} else {
				zlog.Error(fmt.Sprintf("Unknown service type - %s", depReq.ServiceType))
				sendDeploymentResponse(sender, false, "Unknown service type.")
			}
		}
	}

}

func handleCardanoDeployment(depReq models.DeploymentRequest, sender string) {
	// dowload kernel and filesystem files place them somewhere
	// TODO : organize fc files
	kernelFileURL := "https://d.nunet.io/fc/vmlinux"
	kernelFilePath := "/etc/nunet/vmlinux"
	filesystemURL := "https://d.nunet.io/fc/nunet-fc-ubuntu-20.04-0.ext4"
	filesystemPath := "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4"
	pKey := depReq.Params.PublicKey
	nodeId := depReq.Params.NodeID

	err := utils.DownloadFile(kernelFileURL, kernelFilePath)
	if err != nil {
		zlog.Error(fmt.Sprintf("Downloading %s, - %s", kernelFileURL, err.Error()))
		sendDeploymentResponse(sender,
			false,
			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", kernelFileURL))
	}
	err = utils.DownloadFile(filesystemURL, filesystemPath)
	if err != nil {
		zlog.Error(fmt.Sprintf("Downloading %s - %s", filesystemURL, err.Error()))
		sendDeploymentResponse(sender,
			false,
			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", filesystemURL))
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
		zlog.Error(err.Error())
		sendDeploymentResponse(sender, false, "Cardano Node Deployment Failed. Unable to spawn VM.")
	}
	zlog.Error(string(respBody))
}

func handleGpuDeployment(depReq models.DeploymentRequest, sender string) {
	depResp := models.DeploymentResponse{}

	callID := statsdb.GetCallID()
	peerIDOfServiceHost := depReq.Params.NodeID
	timeStamp := float32(statsdb.GetTimestamp())
	status := "accepted"

	ServiceCallParams := models.ServiceCall{
		CallID:              callID,
		PeerIDOfServiceHost: peerIDOfServiceHost,
		ServiceID:           depReq.ServiceType,
		CPUUsed:             float32(depReq.Constraints.CPU),
		MaxRAM:              float32(depReq.Constraints.Vram),
		MemoryUsed:          float32(depReq.Constraints.RAM),
		NetworkBwUsed:       0.0,
		TimeTaken:           0.0,
		Status:              status,
		Timestamp:           timeStamp,
	}
	statsdb.ServiceCall(ServiceCallParams)

	requestTracker := models.RequestTracker{
		ServiceType: depReq.ServiceType,
		CallID:      callID,
		NodeID:      peerIDOfServiceHost,
		Status:      status,
	}
	result := db.DB.Create(&requestTracker)
	if result.Error != nil {
		panic(result.Error)
	}

	depResp = docker.HandleDeployment(depReq, depResp)
	var m map[string]interface{}
	b, _ := json.Marshal(&depResp)
	_ = json.Unmarshal(b, &m)

	var genericMsg models.GenericMessage
	genericMsg.Type = "DeploymentResponse"
	genericMsg.Message = m
	jsonGenericMsg, _ := json.Marshal(genericMsg)
	adapter.SendMessage(sender, string(jsonGenericMsg))
}
