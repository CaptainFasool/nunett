// Package messaging contains logic to listen to queues/channels spread over the application.
package messaging

import (
	"encoding/json"

	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/models"
)

func sendDeploymentResponse(address string, success bool, message string) {
	jsonDepResp, _ := json.Marshal(models.DeploymentResponse{
		Success: success,
		Content: message,
	})
	adapter.SendMessage(address, string(jsonDepResp))
}

func DeploymentWorker() {
	// var adapterMessage models.AdapterMessage

	// err := json.Unmarshal([]byte(messageValue), &adapterMessage)
	// if err != nil {
	// 	fmt.Println("            ", err)
	// 	log.Println("Error: Unable to parse received message: " + err.Error())
	// }

	// if adapterMessage.Message.ServiceType == "cardano_node" {
	// 	// dowload kernel and filesystem files place them somewhere
	// 	// TODO : organize fc files
	// 	kernelFileURL := "https://d.nunet.io/fc/vmlinux"
	// 	kernelFilePath := "/etc/nunet/vmlinux"
	// 	filesystemURL := "https://d.nunet.io/fc/nunet-fc-ubuntu-20.04-0.ext4"
	// 	filesystemPath := "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4"
	// 	pKey := adapterMessage.Message.Params.PublicKey
	// 	nodeId := adapterMessage.Message.Params.NodeID

	// 	err = utils.DownloadFile(kernelFileURL, kernelFilePath)
	// 	if err != nil {
	// 		log.Println("Error: Downloading ", kernelFileURL, " - ", err.Error())
	// 		sendDeploymentResponse(adapterMessage.Sender,
	// 			false,
	// 			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", kernelFileURL))
	// 	}
	// 	err = utils.DownloadFile(filesystemURL, filesystemPath)
	// 	if err != nil {
	// 		log.Println("Error: Downloading ", filesystemURL, " - ", err.Error())
	// 		sendDeploymentResponse(adapterMessage.Sender,
	// 			false,
	// 			fmt.Sprintf("Cardano Node Deployment Failed. Unable to download %s", filesystemURL))
	// 	}

	// 	// makerequest to start-default with downloaded files.
	// 	startDefaultBody := struct {
	// 		KernelImagePath string `json:"kernel_image_path"`
	// 		FilesystemPath  string `json:"filesystem_path"`
	// 		PublicKey       string `json:"public_key"`
	// 		NodeID          string `json:"node_id"`
	// 	}{
	// 		KernelImagePath: kernelFilePath,
	// 		FilesystemPath:  filesystemPath,
	// 		PublicKey:       pKey,
	// 		NodeID:          nodeId,
	// 	}
	// 	jsonBody, _ := json.Marshal(startDefaultBody)

	// 	resp := utils.MakeInternalRequest(&gin.Context{}, "POST", "/vm/start-default", jsonBody)
	// 	respBody, err := io.ReadAll(resp.Body)
	// 	if err != nil {
	// 		log.Println(err)
	// 		sendDeploymentResponse(adapterMessage.Sender, false, "Cardano Node Deployment Failed. Unable to spawn VM.")
	// 	}
	// 	log.Println(string(respBody))

	// } else if adapterMessage.Message.ServiceType == "ml-training-gpu" {
	// 	depResp := models.DeploymentResponse{}

	// 	callID := statsdb.GetCallID()
	// 	peerIDOfServiceHost := adapterMessage.Message.Params.NodeID
	// 	timeStamp := float32(statsdb.GetTimestamp())
	// 	status := "accepted"

	// 	ServiceCallParams := models.ServiceCall{
	// 		CallID:              callID,
	// 		PeerIDOfServiceHost: peerIDOfServiceHost,
	// 		ServiceID:           adapterMessage.Message.ServiceType,
	// 		CPUUsed:             float32(adapterMessage.Message.Constraints.CPU),
	// 		MaxRAM:              float32(adapterMessage.Message.Constraints.Vram),
	// 		MemoryUsed:          float32(adapterMessage.Message.Constraints.RAM),
	// 		NetworkBwUsed:       0.0,
	// 		TimeTaken:           0.0,
	// 		Status:              status,
	// 		Timestamp:           timeStamp,
	// 	}
	// 	statsdb.ServiceCall(ServiceCallParams)

	// 	requestTracker := models.RequestTracker{
	// 		ServiceType: adapterMessage.Message.ServiceType,
	// 		CallID:      callID,
	// 		NodeID:      peerIDOfServiceHost,
	// 		Status:      status,
	// 	}
	// 	result := db.DB.Create(&requestTracker)
	// 	if result.Error != nil {
	// 		panic(result.Error)
	// 	}

	// 	depResp = docker.HandleDeployment(adapterMessage.Message, depResp)
	// 	jsonDepResp, _ := json.Marshal(depResp)
	// 	SendMessage(adapterMessage.Sender, string(jsonDepResp))
	// } else {
	// 	log.Println("Error: Uknown service type - ", adapterMessage.Message.ServiceType)
	// 	sendDeploymentResponse(adapterMessage.Sender, false, "Unknown service type.")
	// }
}
