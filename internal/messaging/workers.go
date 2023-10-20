// Package messaging contains logic to listen to queues/channels spread over the application.
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/docker"
	elk "gitlab.com/nunet/device-management-service/internal/heartbeat"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

// GetCallID returns a call ID to track the deployement request
func GetCallID() int64 {
	min := int64(1e15)
	max := int64(1e16 - 1)
	return min + rand.Int63n(max-min+1)
}

func sendDeploymentResponse(success bool, content string) {
	depResp, _ := json.Marshal(&models.DeploymentResponse{
		Success: success,
		Content: content,
	})

	zlog.Sugar().Debugf("marshalled deployment response: %s", string(depResp))

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
			ctx := context.Background()
			var depReq models.DeploymentRequest
			jsonDataMsg, _ := json.Marshal(msg)
			json.Unmarshal(jsonDataMsg, &depReq)

			if depReq.ServiceType == "cardano_node" {
				handleCardanoDeployment(depReq)
			} else if depReq.ServiceType == "ml-training-cpu" || depReq.ServiceType == "ml-training-gpu" {
				handleDockerDeployment(ctx, depReq)
			} else {
				zlog.Error(fmt.Sprintf("Unknown service type - %s", depReq.ServiceType))
				sendDeploymentResponse(false, "Unknown service type.")
			}
		}
	}
}

// FileTransferWorker continuously listens to the file transfer queue and
// filters which files to accept.
func FileTransferWorker() {
	for msg := range libp2p.FileTransferQueue {
		// implement your own logic which files should be accepted
		// Check 1: The sender peer must be the same peer with whom deployment request stream is open
		// if no stream is open, reject the file
		if msg.Sender != libp2p.OutboundDepReqStream.Conn().RemotePeer() {
			zlog.Sugar().Errorf("File Transfer Request from %s rejected. No deployment request stream open.", msg.Sender)
			continue
		}

		// Check 2: The file must end with tar.gz or tar.gz.sha256.txt
		if !(strings.HasSuffix(msg.File.Name, "tar.gz") || strings.HasSuffix(msg.File.Name, "tar.gz.sha256.txt")) {
			zlog.Sugar().Errorf("File Transfer Request from %s rejected. File type not supported (%s).", msg.Sender, msg.File.Name)
			continue
		}

		// accept the file
		zlog.Sugar().Infof("File Transfer Request of %s from %s accepted.", msg.File.Name, msg.Sender)
		go libp2p.AcceptFileTransfer()
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

	resp, err := utils.MakeInternalRequest(&gin.Context{}, "POST", "/api/v1/vm/start-default", jsonBody)
	if err != nil {
		zlog.Error(err.Error())
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		zlog.Error(err.Error())
		sendDeploymentResponse(false, "Cardano Node Deployment Failed. Unable to spawn VM.")
		return
	}
	sendDeploymentResponse(true, "Cardano Node Deployment Successful.")
}

func handleDockerDeployment(ctx context.Context, depReq models.DeploymentRequest) {
	depResp := models.DeploymentResponse{}
	callID := GetCallID()
	peerIDOfServiceHost := depReq.Params.LocalNodeID
	status := "accepted"

	requestTracker := models.RequestTracker{
		ID:          1,
		ServiceType: depReq.ServiceType,
		CallID:      callID,
		NodeID:      peerIDOfServiceHost,
		Status:      status,
		MaxTokens:   depReq.MaxNtx,
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

	// sending service call info to elastic search
	elk.ProcessUsage(int(callID), 0, 0, 0, 0, depReq.MaxNtx)

	depResp = docker.HandleDeployment(ctx, depReq)
	sendDeploymentResponse(depResp.Success, depResp.Content)
}
