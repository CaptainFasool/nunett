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
	"gitlab.com/nunet/device-management-service/statsdb"
	"gitlab.com/nunet/device-management-service/utils"
)

func sendDeploymentResponse(success bool, content string) {

	zlog.Sugar().Debugf("send deployment response content: %s", content)

	depResp, _ := json.Marshal(&models.DeploymentResponse{
		Success: success,
		Content: content,
	})

	zlog.Sugar().Debugf("marshalled deployment response from worker: %s", string(depResp))

	var closeStream bool
	if !success {
		closeStream = true
	}

	err := libp2p.DeploymentUpdate(libp2p.MsgDepResp, string(depResp), closeStream)
	if err != nil {
		zlog.Sugar().Errorf("Error Sending Deployment Response - ", err.Error())
	}
}

func DeploymentWorker() {
	for {
		select {
		case msg := <-libp2p.DepReqQueue:
			var depReq models.DeploymentRequest

			jsonDataMsg, _ := json.Marshal(msg)
			json.Unmarshal(jsonDataMsg, &depReq)

			if depReq.ServiceType == "cardano_node" {
				handleCardanoDeployment(depReq)
			} else if depReq.ServiceType == "ml-training-cpu" || depReq.ServiceType == "ml-training-gpu" {
				handleDockerDeployment(depReq)
			} else {
				zlog.Error(fmt.Sprintf("Unknown service type - %s", depReq.ServiceType))
				sendDeploymentResponse(false, "Unknown service type.")
			}
		}
	}
}

func handleCardanoDeployment(depReq models.DeploymentRequest) {
	// dowload kernel and filesystem files place them somewhere
	// TODO : organize fc files
	pKey := depReq.Params.LocalPublicKey
	nodeId := depReq.Params.LocalNodeID

	err := utils.DownloadFile(utils.KernelFileURL, utils.KernelFilePath)
	if err != nil {
		zlog.Error(fmt.Sprintf("Downloading %s, - %s", utils.KernelFileURL, err.Error()))
		sendDeploymentResponse(false,
			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", utils.KernelFileURL))
		return
	}
	err = utils.DownloadFile(utils.FilesystemURL, utils.FilesystemPath)
	if err != nil {
		zlog.Error(fmt.Sprintf("Downloading %s - %s", utils.FilesystemURL, err.Error()))
		sendDeploymentResponse(false,
			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", utils.FilesystemURL))
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
		sendDeploymentResponse(false, "Cardano Node Deployment Failed. Unable to spawn VM.")
		return
	}
	sendDeploymentResponse(true, "Cardano Node Deployment Successful.")
}

func handleDockerDeployment(depReq models.DeploymentRequest) {
	depResp := models.DeploymentResponse{}
	callID := statsdb.GetCallID()
	peerIDOfServiceHost := depReq.Params.LocalNodeID
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
		ID:          1,
		ServiceType: depReq.ServiceType,
		CallID:      callID,
		NodeID:      peerIDOfServiceHost,
		Status:      status,
		MaxTokens:   int32(depReq.MaxNtx),
	}

	// Check if we have a previous entry in the table
	var reqTracker models.RequestTracker
	if res := db.DB.Find(&reqTracker); res.RowsAffected == 0 {
		result := db.DB.Create(&requestTracker)
		if result.Error != nil {
			zlog.Error(result.Error.Error())
		}
	} else {
		result := db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Select("*").Updates(requestTracker)
		if result.Error != nil {
			zlog.Error(result.Error.Error())
		}
	}

	depResp = docker.HandleDeployment(depReq)
	sendDeploymentResponse(depResp.Success, depResp.Content)
}
