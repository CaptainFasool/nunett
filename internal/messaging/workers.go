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
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	elk "gitlab.com/nunet/device-management-service/telemetry/heartbeat"
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

			if depReq.Params.ResumeJob.Resume {
				depReq.Params.ResumeJob.ProgressFile = <-ProgressFilePathOnCP
				zlog.Debug("file transfer likely to complete by now")
			}

			if depReq.ServiceType == "cardano_node" {
				content, err := handleCardanoDeployment(depReq)
				if err != nil {
					zlog.Sugar().Errorln(err)
					sendDeploymentResponse(false, content)
				}
				sendDeploymentResponse(true, content)
			} else if depReq.ServiceType == "ml-training-cpu" || depReq.ServiceType == "ml-training-gpu" {
				depResp, err := handleDockerDeployment(ctx, depReq)
				if err != nil {
					zlog.Sugar().Errorln(err)
					sendDeploymentResponse(false, err.Error())
				}
				sendDeploymentResponse(true, depResp.Content)
			} else {
				zlog.Error(fmt.Sprintf("Unknown service type - %s", depReq.ServiceType))
				sendDeploymentResponse(false, "Unknown service type.")
			}
		}
	}
}

// FileTransferWorker continuously listens to the file transfer queue and
// filters which files to accept.
func FileTransferWorker(ctx context.Context) {
	for msg := range libp2p.FileTransferQueue {
		// Check if there's an ongoing deployment request stream (either inbound or outbound) with the sender
		if (libp2p.InboundDepReqStream != nil && libp2p.InboundDepReqStream.Conn().RemotePeer() == msg.Sender) ||
			(libp2p.OutboundDepReqStream != nil && libp2p.OutboundDepReqStream.Conn().RemotePeer() == msg.Sender) {
			// Check if the file type is supported (e.g., ends with "tar.gz")
			if strings.HasSuffix(msg.File.Name, "tar.gz") {
				zlog.Info(fmt.Sprintf("File Transfer Request of file %q from %s accepted.", msg.File.Name, msg.Sender))

				resultChan := make(chan libp2p.FileTransferResult)
				go func() {
					filePath, transferChan, err := libp2p.AcceptFileTransfer(ctx, msg)
					resultChan <- libp2p.FileTransferResult{
						FilePath:     filePath,
						TransferChan: transferChan,
						Error:        err,
					}
				}()

				result := <-resultChan
				// wait for channel to be closed (meaning file transfer is 100%)
				<-result.TransferChan
				zlog.Info("File transfer complete!")
				ProgressFilePathOnCP <- result.FilePath
				zlog.Debug("Deployment request should proceed now")

				if strings.HasSuffix(msg.File.Name, "tar.gz") {
					// verify checksum before extraction
					zlog.Sugar().Infof("Verifying SHA-256 checksum of %s", result.FilePath)
					checksum, err := utils.CalculateSHA256Checksum(result.FilePath)
					if err != nil {
						zlog.Sugar().Errorf("Error calculating SHA-256 checksum of %s - %s", result.FilePath, err.Error())
					}
					if checksum != msg.File.SHA256Checksum {
						zlog.Sugar().Errorf("SHA-256 checksum of %s does not match expected checksum %s", result.FilePath, msg.File.SHA256Checksum)
						//XXX should we delete the file if it's not the right file?
						return
					}

					// write checksum to file
					sha256FilePath, err := utils.CreateCheckSumFile(result.FilePath, checksum)
					if err != nil {
						zlog.Sugar().Errorf("Error writing SHA-256 checksum to file - %v", err)
					}

					zlog.Sugar().Debugf("Checksum matched! SHA-256 written to %q.", sha256FilePath)
				}
			} else {
				zlog.Error(fmt.Sprintf("File Transfer Request from %s rejected. File type not supported.", msg.Sender))
			}
		} else {
			zlog.Error(fmt.Sprintf("File Transfer Request from %s rejected. No deployment request stream open with the sender.", msg.Sender))
		}
	}
}

func handleCardanoDeployment(depReq models.DeploymentRequest) (string, error) {
	// dowload kernel and filesystem files place them somewhere
	// TODO : organize fc files
	pKey := depReq.Params.LocalPublicKey
	nodeId := depReq.Params.LocalNodeID

	err := utils.DownloadFile(utils.KernelFileURL, utils.KernelFilePath)
	if err != nil {
		content := fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", utils.KernelFileURL)
		return content, fmt.Errorf(fmt.Sprintf("Downloading %s, - %s", utils.KernelFileURL, err.Error()))
	}
	err = utils.DownloadFile(utils.FilesystemURL, utils.FilesystemPath)
	if err != nil {
		content := fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", utils.FilesystemURL)
		return content, fmt.Errorf(fmt.Sprintf("Downloading %s - %s", utils.FilesystemURL, err.Error()))
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

	resp, err := utils.MakeInternalRequest(&gin.Context{}, "POST", "/api/v1/vm/start-default", "", jsonBody)
	if err != nil {
		zlog.Error(err.Error())
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		content := "Cardano Node Deployment Failed. Unable to spawn VM."
		return content, fmt.Errorf(err.Error())
	}
	return "Cardano Node Deployment Successful.", nil
}

func handleDockerDeployment(ctx context.Context, depReq models.DeploymentRequest) (models.DeploymentResponse, error) {
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
			return models.DeploymentResponse{}, fmt.Errorf("failed to read requestTracker table: %v", result.Error.Error())
		}
	} else {
		result := db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Select("*").Updates(requestTracker)
		if result.Error != nil {
			return models.DeploymentResponse{}, fmt.Errorf("failed to update requestTracker table: %v", result.Error.Error())
		}
	}

	// sending service call info to elastic search
	err := elk.ProcessUsage(int(callID), 0, 0, 0, 0, depReq.MaxNtx)
	if err != nil {
		return models.DeploymentResponse{}, err
	}

	depResp, err = docker.HandleDeployment(ctx, depReq)
	if err != nil {
		return models.DeploymentResponse{}, err
	}

	return depResp, nil
}
