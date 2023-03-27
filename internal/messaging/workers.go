// Package messaging contains logic to listen to queues/channels spread over the application.
package messaging

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/docker"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"

	//"gitlab.com/nunet/device-management-service/statsdb" //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
	"gitlab.com/nunet/device-management-service/utils"
)

func sendDeploymentResponse(success bool, message string, close bool) {
	jsonDepResp, _ := json.Marshal(models.DeploymentResponse{
		Success: success,
		Content: message,
	})
	err := libp2p.DeploymentResponse(string(jsonDepResp), close)
	if err != nil {
		zlog.Sugar().Errorln("Error Sending Deployment Response - ", err.Error())
	}
}

func DeploymentWorker() {
	for {
		select {
		case msg := <-libp2p.DepReqQueue:
			var depReq models.DeploymentRequest

			jsonDataMsg, _ := json.Marshal(msg)
			json.Unmarshal(jsonDataMsg, &depReq)

			sender := depReq.AddressUser
			if depReq.ServiceType == "cardano_node" {
				handleCardanoDeployment(depReq, sender)
			} else if depReq.ServiceType == "ml-training-gpu" {
				handleGpuDeployment(depReq)
			} else {
				zlog.Error(fmt.Sprintf("Unknown service type - %s", depReq.ServiceType))
				sendDeploymentResponse(false, "Unknown service type.", true)
			}
		}
	}
}

func handleCardanoDeployment(depReq models.DeploymentRequest, sender string) {
	// dowload kernel and filesystem files place them somewhere
	// TODO : organize fc files
	pKey := depReq.Params.PublicKey
	nodeId := depReq.Params.NodeID

	err := utils.DownloadFile(utils.KernelFileURL, utils.KernelFilePath)
	if err != nil {
		zlog.Error(fmt.Sprintf("Downloading %s, - %s", utils.KernelFileURL, err.Error()))
		sendDeploymentResponse(false,
			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", utils.KernelFileURL), true)
		return
	}
	err = utils.DownloadFile(utils.FilesystemURL, utils.FilesystemPath)
	if err != nil {
		zlog.Error(fmt.Sprintf("Downloading %s - %s", utils.FilesystemURL, err.Error()))
		sendDeploymentResponse(false,
			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", utils.FilesystemURL), true)
		return
	}

	// makerequest to start-default with downloaded files.
	startDefaultBody := struct {
		KernelImagePath string `json:"kernel_image_path"`
		FilesystemPath  string `json:"filesystem_path"`
		PublicKey       string `json:"public_key"`
		NodeID          string `json:"node_id"`
	}{
		KernelImagePath: utils.KernelFilePath,
		FilesystemPath:  utils.FilesystemPath,
		PublicKey:       pKey,
		NodeID:          nodeId,
	}
	jsonBody, _ := json.Marshal(startDefaultBody)

	resp := utils.MakeInternalRequest(&gin.Context{}, "POST", "/vm/start-default", jsonBody)
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		zlog.Error(err.Error())
		sendDeploymentResponse(false, "Cardano Node Deployment Failed. Unable to spawn VM.", true)
		return
	}
	sendDeploymentResponse(true, "Cardano Node Deployment Successful.", false)
}

func handleGpuDeployment(depReq models.DeploymentRequest) {
	depResp := models.DeploymentResponse{}

	callID := float32(1234) //statsdb.GetCallID() //XXX: Using dummy value until StatsDB works - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
	peerIDOfServiceHost := depReq.Params.NodeID
	//timeStamp := float32(statsdb.GetTimestamp()) //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
	status := "accepted"

	// ServiceCallParams := models.ServiceCall{ //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
	// 	CallID:              callID,
	// 	PeerIDOfServiceHost: peerIDOfServiceHost,
	// 	ServiceID:           depReq.ServiceType,
	// 	CPUUsed:             float32(depReq.Constraints.CPU),
	// 	MaxRAM:              float32(depReq.Constraints.Vram),
	// 	MemoryUsed:          float32(depReq.Constraints.RAM),
	// 	NetworkBwUsed:       0.0,
	// 	TimeTaken:           0.0,
	// 	Status:              status,
	// 	Timestamp:           timeStamp,
	// }
	// statsdb.ServiceCall(ServiceCallParams) //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138

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

	jsonGenericMsg, _ := json.Marshal(m)
	sendDeploymentResponse(depResp.Success, string(jsonGenericMsg), false)
}
