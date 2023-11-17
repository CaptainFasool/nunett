package docker

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/KyleBanks/dockerstats"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/libp2p/go-libp2p/core/peer"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/internal/config"
	elk "gitlab.com/nunet/device-management-service/internal/heartbeat"
	library "gitlab.com/nunet/device-management-service/lib"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
	"gitlab.com/nunet/device-management-service/telemetry"
	"gitlab.com/nunet/device-management-service/utils"
	"go.uber.org/zap"
)

var (
	cpuPeriod         int64         = 100000
	logUpdateInterval time.Duration = time.Duration(config.GetConfig().Job.LogUpdateInterval) * time.Minute
)

func freeUsedResources() {
	// update the available resources table
	err := telemetry.CalcFreeResAndUpdateDB()
	if err != nil {
		zlog.Sugar().Errorf("Error calculating and updating FreeResources: %v", err)
	}
	freeResource, err := telemetry.GetFreeResources()
	if err != nil {
		zlog.Sugar().Errorf("Error getting freeResources: %v", err)
	}
	elk.DeviceResourceChange(freeResource.TotCpuHz, freeResource.Ram)
	libp2p.UpdateKadDHT()
}

func mhzPerCore() (float64, error) {
	// cpus, err := cpu.Info()
	// if err != nil {
	// 	zlog.Sugar().Errorf("failed to get cpu info: %v", err)
	// 	return 0, err
	// }
	return library.Hz_per_cpu(), nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func mhzToVCPU(cpuInMhz int) (float64, error) {
	mhz, err := mhzPerCore()
	if err != nil {
		return 0, err
	}
	vcpu := float64(cpuInMhz) / mhz
	return toFixed(vcpu, 2), nil
}

func groupExists(groupName string) bool {
	_, err := user.LookupGroup(groupName)
	return err == nil
}

// RunContainer goes through the process of setting constraints,
// specifying image name and cmd. It starts a container and posts
// log update every logUpdateDuration.
func RunContainer(ctx context.Context, depReq models.DeploymentRequest, createdLog LogbinResponse, resCh chan<- models.DeploymentResponse, servicePK uint, chosenGPUVendor library.GPUVendor) {
	zlog.Info("Entering RunContainer")
	machine_type := depReq.Params.MachineType
	gpuOpts := opts.GpuOpts{}
	if machine_type == "gpu" {
		gpuOpts.Set("all") // TODO find a way to use GPU and CPU
	}
	imageName := depReq.Params.ImageID
	if chosenGPUVendor == library.AMD {
		imageName += "-amd"
	}
	modelURL := depReq.Params.ModelURL
	packages := strings.Join(depReq.Params.Packages, " ")
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   []string{modelURL, packages},
	}
	// Get onboarded resources
	cpuQuota, memoryMax, err := fetchOnboardedResources()
	if err != nil {
		zlog.Sugar().Errorf("Error fetching onboarded resources: %v", err)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with onboarded resources. Unable to process request."}
		resCh <- depRes
		return
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			DeviceRequests: gpuOpts.Value(),
			Memory:         memoryMax,
			CPUQuota:       cpuQuota,
			CPUPeriod:      cpuPeriod,
		},
	}

	if depReq.Params.Container.PortToBind != "" {
		natPortToBind, err := nat.NewPort("tcp", depReq.Params.Container.PortToBind)
		if err != nil {
			zlog.Sugar().Errorf("Error creating port to bind: %v", err)
			depRes := models.DeploymentResponse{
				Success: false,
				Content: "Problem with port to bind. Unable to process request.",
			}
			resCh <- depRes
			return
		}

		vpnAddr, err := libp2p.GetVPNAddrOfHost()
		if err != nil {
			zlog.Sugar().Errorf("Error getting VPN address: %v", err)
			depRes := models.DeploymentResponse{
				Success: false,
				Content: "Problem with VPN address. Unable to process request.",
			}
			resCh <- depRes
			return
		}

		hostConfig.PortBindings = nat.PortMap{
			natPortToBind: []nat.PortBinding{
				{
					HostIP:   vpnAddr,
					HostPort: depReq.Params.Container.PortToBind,
				},
			},
		}
	}

	hostConfigAMDGPU := container.HostConfig{}

	if chosenGPUVendor == library.AMD {
		hostConfigAMDGPU = container.HostConfig{
			Binds: []string{
				"/dev/kfd:/dev/kfd",
				"/dev/dri:/dev/dri",
			},
			Resources: container.Resources{
				Memory:    memoryMax,
				CPUQuota:  cpuQuota,
				CPUPeriod: cpuPeriod,
				Devices: []container.DeviceMapping{
					{
						PathOnHost:        "/dev/kfd",
						PathInContainer:   "/dev/kfd",
						CgroupPermissions: "rwm",
					},
					{
						PathOnHost:        "/dev/dri",
						PathInContainer:   "/dev/dri",
						CgroupPermissions: "rwm",
					},
				},
			},
			GroupAdd: []string{"video"},
		}

		// For Ubuntu > 18.04
		if groupExists("render") {
			hostConfigAMDGPU.GroupAdd = append(hostConfigAMDGPU.GroupAdd, "render")
		}

		hostConfig = &hostConfigAMDGPU
	}

	var freeRes models.FreeResources

	if res := db.DB.Find(&freeRes); res.RowsAffected == 0 {
		zlog.Sugar().Errorf("Record not found!")
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with Free Resources. Unable to process request."}
		resCh <- depRes
		return
	}

	// Check if we have enough free resources before running Container
	if (depReq.Constraints.RAM > freeRes.Ram) ||
		(depReq.Constraints.CPU > freeRes.TotCpuHz) {
		zlog.Sugar().Errorf("Not enough resources available to deploy container")
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with resources for deployment. Unable to process request."}
		resCh <- depRes
		return
	}

	// if resuming a job, bind the volume together
	if depReq.Params.ResumeJob.Resume {
		hostPath := filepath.Join(config.GetConfig().General.MetadataPath, "progress")
		containerPath := "/workspace"
		volumeBinding := fmt.Sprintf("%s:%s", hostPath, containerPath)

		// preserve other binds from AMD host config
		if chosenGPUVendor == library.AMD {
			hostConfig.Binds = append(hostConfig.Binds, volumeBinding)
		} else {
			hostConfig = &container.HostConfig{
				Binds: []string{volumeBinding},
			}
		}
	}

	resp, err := dc.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		zlog.Sugar().Errorf("Unable to create container: %v", err)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with container. Unable to process request."}
		resCh <- depRes
		return
	}

	if err := dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		zlog.Sugar().Errorf("Unable to start container: %v", err)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with running container. Unable to process request."}
		resCh <- depRes
		return
	}

	var requestTracker models.RequestTracker
	res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	}
	status := "started"

	// updating status of service to elastic search
	elk.ProcessStatus(int(requestTracker.CallID), requestTracker.NodeID, "", status, 0)

	// updating RequestTracker
	requestTracker.Status = status
	requestTracker.RequestID = resp.ID
	res = db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Updates(requestTracker)
	if res.Error != nil {
		zlog.Sugar().Errorf("unable to update request tracker: %v", res.Error)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with request tracker. Unable to process request."}
		resCh <- depRes
		return
	}

	var resourceRequirements models.ServiceResourceRequirements
	resourceRequirements.CPU = depReq.Constraints.CPU
	resourceRequirements.RAM = depReq.Constraints.RAM

	result := db.DB.Create(&resourceRequirements)
	if result.Error != nil {
		zlog.Sugar().Errorf("unable to create resource requirements: %v", res.Error)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with resource requirements. Unable to process request."}
		resCh <- depRes
		return
	}

	// Update db - find the service based on primary key and update container id
	var service models.Services
	res = db.DB.Model(&service).Where("id = ?", servicePK).Updates(models.Services{ContainerID: resp.ID, ResourceRequirements: int(resourceRequirements.ID)})
	if res.Error != nil {
		zlog.Sugar().Errorf("unable to update services: %v", res.Error)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with services tracker. Unable to process request."}
		resCh <- depRes
		return
	}

	err = telemetry.CalcFreeResAndUpdateDB()
	if err != nil {
		zlog.Sugar().Errorf("Error calculating and updating FreeResources: %v", err)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with free resources calculation. Unable to process request."}
		resCh <- depRes
		return
	}

	libp2p.UpdateKadDHT()

	depRes := models.DeploymentResponse{Success: true}
	resCh <- depRes

	tick := time.NewTicker(logUpdateInterval)
	defer tick.Stop()

	containerTimeout := time.NewTimer(time.Duration(depReq.Constraints.Time) * time.Minute)
	defer containerTimeout.Stop()

	statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	maxUsedRAM, maxUsedCPU, networkBwUsed := 0.0, 0.0, 0.0

outerLoop:
	for {
		select {
		case err := <-errCh:
			zlog.Info("[container running] entering first case; container didn't start")
			// handle error & exit
			if err != nil {
				zlog.Sugar().Errorf("problem in running contianer: %v", err)
				depRes := models.DeploymentResponse{Success: false, Content: "Problem occurred with container. Unable to complete request."}
				resCh <- depRes
				freeUsedResources()
				return
			}
		case containerStatus := <-statusCh: // not running?
			zlog.Info("[container running] entering second case; container exiting")

			// get the last logs & exit...
			updateLogbin(ctx, createdLog.ID, resp.ID)

			// Add a response for log update
			if r := db.DB.Where("container_id = ?", resp.ID).First(&service); r.Error != nil {
				zlog.Sugar().Errorf("problem updating services: %v", r.Error)
				service.JobStatus = "finished with errors"
			}
			sendLogsToSPD(ctx, resp.ID, service.LastLogFetch.Format("2006-01-02T15:04:05Z"))

			// add exitStatus to db
			if containerStatus.StatusCode == 0 {
				service.JobStatus = libp2p.ContainerJobFinishedWithoutErrors
				status = libp2p.ContainerJobFinishedWithoutErrors
				requestTracker.Status = libp2p.ContainerJobFinishedWithoutErrors
			} else if containerStatus.StatusCode > 0 {
				service.JobStatus = libp2p.ContainerJobFinishedWithErrors
				status = libp2p.ContainerJobFinishedWithErrors
				requestTracker.Status = libp2p.ContainerJobFinishedWithErrors
			}

			duration := time.Now().Sub(service.CreatedAt)
			jobDuration := duration.Minutes()
			service.JobDuration = int64(jobDuration)

			oracleResp := getSignaturesFromOracle(service) // saving the signatures into DB, received from Oracle + wallet server
			service.TransactionType = oracleResp.RewardType
			service.SignatureDatum = oracleResp.SignatureDatum
			service.MessageHashDatum = oracleResp.MessageHashDatum
			service.Datum = oracleResp.Datum
			service.SignatureAction = oracleResp.SignatureAction
			service.MessageHashAction = oracleResp.MessageHashAction
			service.Action = oracleResp.Action
			db.DB.Save(&service)

			// Send service status update
			serviceBytes, _ := json.Marshal(service)
			var closeStream bool
			if strings.Contains(string(service.JobStatus), "finished") {
				closeStream = true
			}
			libp2p.DeploymentUpdate(libp2p.MsgJobStatus, string(serviceBytes), closeStream)

			timeTaken := duration.Seconds()
			averageNetBw := networkBwUsed / timeTaken

			// updating elastic search with status of service
			elk.ProcessUsage(int(requestTracker.CallID), int(maxUsedCPU), int(maxUsedRAM),
				int(averageNetBw), int(timeTaken), requestTracker.MaxTokens)
			elk.ProcessStatus(int(requestTracker.CallID), depReq.Params.LocalNodeID, "", status, 0)
			db.DB.Save(requestTracker)
			freeUsedResources()
			break outerLoop
		case <-tick.C:
			zlog.Info("[container running] entering third case; time ticker")

			// get the latest logs ...
			contID := requestTracker.RequestID[:12]
			stats, err := dockerstats.Current()
			if err != nil {
				zlog.Sugar().Errorf("problem obtaining docker stats: %v", err)
			}

			var tempService models.Services
			if err := db.DB.Where("container_id = ?", resp.ID).First(&tempService).Error; err != nil {
				panic(err)
			}
			duration := time.Since(tempService.CreatedAt)
			timeTaken := duration.Seconds()
			averageNetBw := networkBwUsed / timeTaken

			for _, s := range stats {
				if s.Container == contID {
					usedRAM := strings.Split(s.Memory.Raw, "/")
					usedCPU := strings.Split(s.CPU, "%")
					usedNetwork := strings.Split(s.IO.Network, "/")
					ramFloat64 := calculateResourceUsage(usedRAM[0])
					cpuFloat64, _ := strconv.ParseFloat(usedCPU[0], 64)
					cpuFloat64 = cpuUsage(cpuFloat64, float64(depReq.Constraints.CPU))
					networkFloat64Pre := calculateResourceUsage(usedNetwork[0])
					networkFloat64Suf := calculateResourceUsage(usedNetwork[1])
					networkFloat64 := networkFloat64Pre + networkFloat64Suf
					if ramFloat64 > maxUsedRAM {
						maxUsedRAM = ramFloat64 / 1024
					}
					if cpuFloat64 > maxUsedCPU {
						maxUsedCPU = cpuFloat64
					}
					if networkFloat64 > networkBwUsed {
						networkBwUsed = networkFloat64
					}
				}
			}

			// updating elastic search with telemetry info during job running
			elk.ProcessUsage(int(requestTracker.CallID), int(maxUsedCPU), int(maxUsedRAM),
				int(averageNetBw), int(timeTaken), requestTracker.MaxTokens)

			updateLogbin(ctx, createdLog.ID, resp.ID)

			// Add a response for log update
			db.DB.Where("container_id = ?", resp.ID).First(&service)
			zlog.Debug("service.LastLogFetch",
				zap.String("value", service.LastLogFetch.Format("2006-01-02T15:04:05Z")),
			)
			sendLogsToSPD(ctx, resp.ID, service.LastLogFetch.Format("2006-01-02T15:04:05Z"))
			service.LastLogFetch = time.Now().In(time.UTC)
			db.DB.Save(&service)

			createTarballAndChecksum(dc, resp.ID, requestTracker.CallID)
		case <-containerTimeout.C:
			// We've hit a case where container has to be forcefully stopped due to timeout
			zlog.Info("[container running] entering fourth case; container timeout")
			beforeContainerTimeout(dc, resp.ID, requestTracker.CallID, depReq.Params.LocalNodeID)
			dc.ContainerStop(ctx, resp.ID, nil)

		}
	}
}

// SearchContianersByImage gets all containers given an imageID string.
func SearchContianersByImage(ctx context.Context, imageID string) ([]types.Container, error) {
	containers, err := dc.ContainerList(ctx, types.ContainerListOptions{
		All: true, Filters: filters.NewArgs(filters.Arg("ancestor", imageID)),
	})
	if err != nil {
		zlog.Sugar().Errorf("unable to find container with imageID : %v", imageID)
	}

	return containers, err
}

// SearchImagesByRefrence gets all container images given a reference string.
// The refrence string should be a regex compilable pattern that will be searched
// against image name (RepoTags) and digests.
func SearchImagesByRefrence(ctx context.Context, reference string) ([]types.ImageSummary, error) {
	var result []types.ImageSummary
	images, err := dc.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		zlog.Sugar().Errorf("unable to find images with reference : %v", reference)
	}
	for _, image := range images {
		refrenceRegex := regexp.MustCompile(reference)
		match := false
		for _, tag := range image.RepoTags {
			if refrenceRegex.MatchString(tag) {
				match = true
			}
		}
		for _, digest := range image.RepoDigests {
			if refrenceRegex.MatchString(digest) {
				match = true
			}
		}
		if match {
			result = append(result, image)
		}
	}
	return result, err
}

// StopAndRemoveContainer stops and removes a container given its ID
func StopAndRemoveContainer(ctx context.Context, containerID string) error {
	if err := dc.ContainerStop(ctx, containerID, nil); err != nil {
		zlog.Sugar().Errorf("Unable to stop container %s: %s", containerID, err)
		return err
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := dc.ContainerRemove(ctx, containerID, removeOptions); err != nil {
		zlog.Sugar().Errorf("Unable to remove container %s: %s", containerID, err)
		return err
	}

	return nil
}

// PullImage is a wrapper around Docker SDK's function with same name.
func PullImage(ctx context.Context, imageName string) error {
	out, err := dc.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		zlog.Sugar().Errorf("unable to pull image: %v", err)
		return err
	}

	type Resp struct {
		Status string `json:"status"`
	}

	defer out.Close()
	d := json.NewDecoder(io.TeeReader(out, os.Stdout))

	var resp *Resp
	var digest string
	newImage := false
	for {
		if err := d.Decode(&resp); err != nil {
			if err == io.EOF {
				break
			}
			zlog.Sugar().Errorf("unable pull image: %v", err)
			return err
		}
		if strings.HasPrefix(resp.Status, "Digest") {
			digest = strings.TrimPrefix(resp.Status, "Digest: ")
		}
		if strings.HasPrefix(resp.Status, "Status: Downloaded") {
			newImage = true
		}
	}

	if newImage {
		images, err := SearchImagesByRefrence(ctx, fmt.Sprintf(".*\\@%v", digest))
		if err != nil || len(images) <= 0 {
			zlog.Sugar().Warnf("unable to find image: %v", imageName)
			return err
		}
		image := images[0]
		imageInfo := models.ContainerImages{ImageID: image.ID, ImageName: imageName, Digest: digest}
		db.DB.Save(&imageInfo)
	}

	return nil
}

func GetContainersFromImage(ctx context.Context, imageID string) ([]types.Container, error) {
	containers, err := dc.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("ancestor", imageID),
		),
	})
	if err != nil {
		zlog.Sugar().Errorf("unable to list containers that use the image: %v\n", imageID)
	}
	return containers, err
}

func RemoveImage(ctx context.Context, imageID string) error {
	_, err := dc.ImageRemove(ctx, imageID, types.ImageRemoveOptions{})
	if err != nil {
		zlog.Sugar().Errorf("unable to remove image: %v\n", imageID)
	}
	return err
}

func cpuUsage(cpu float64, maxCPU float64) float64 {
	return maxCPU * cpu / 100
}

func extractResourceUsage(input string) (float64, string) {
	re := regexp.MustCompile(`(\d+(\.\d+)?)([KkMmGgTt][Bb]|[KkMmGgTt][Ii]?[Bb])`)
	matches := re.FindStringSubmatch(input)
	valueStr := matches[1]
	unit := matches[3]

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0.0, unit
	}

	return value, unit
}

func calculateResourceUsage(input string) float64 {
	value, unit := extractResourceUsage(input)
	switch strings.ToLower(unit) {
	case "kb", "kib":
		return value
	case "mb", "mib":
		return value * 1024
	case "gb", "gib":
		return value * 1024 * 1024
	default:
		return 0.0
	}
}

func getSignaturesFromOracle(service models.Services) (oracleResp *oracle.RewardResponse) {
	oracleResp, err := oracle.Oracle.WithdrawTokenRequest(&oracle.RewardRequest{
		JobStatus:            service.JobStatus,
		JobDuration:          service.JobDuration,
		EstimatedJobDuration: service.EstimatedJobDuration,
		LogPath:              service.LogURL,
		MetadataHash:         service.MetadataHash,
		WithdrawHash:         service.WithdrawHash,
		RefundHash:           service.RefundHash,
		Distribute_50Hash:    service.Distribute_50Hash,
		Distribute_75Hash:    service.Distribute_75Hash,
	})
	if err != nil {
		zlog.Sugar().Errorf("connetction to oracle failed : %v", err)
	}

	return oracleResp
}

// HandleDeployment does following docker based actions in the sequence:
// Pull image, run container, get logs, send log to the requester
func HandleDeployment(ctx context.Context, depReq models.DeploymentRequest) models.DeploymentResponse {
	var chosenGPUVendor library.GPUVendor
	if depReq.Params.MachineType == "gpu" {
		// Finding the GPU with the highest free VRAM regardless of vendor type
		// Get AMD GPU info
		// var gpu_infos [][]library.GPUInfo
		gpu_infos, err := library.GetGPUInfo()
		if err != nil {
			zlog.Sugar().Errorf("GPU/Driver not found: %v", err)
		}
		amdGPUs := gpu_infos[0]
		nvidiaGPUs := gpu_infos[1]

		// amdGPUs, err := library.GetAMDGPUInfo()
		// if err != nil {
		// 	zlog.Sugar().Errorf("AMD GPU/Driver not found: %v", err)
		// }

		// // Get NVIDIA GPU info
		// nvidiaGPUs, err := gpuinfo.GetNVIDIAGPUInfo()
		// if err != nil {
		// 	zlog.Sugar().Errorf("NVIDIA GPU/Driver not found: %v", err)
		// 	// return here and not above for AMD because we need to have at least one GPU
		// 	return models.DeploymentResponse{Success: false, Content: "Unable to get GPU info."}
		// }

		// Combine AMD and NVIDIA GPU info
		allGPUs := append(amdGPUs, nvidiaGPUs...)

		// Find the GPU with the highest free VRAM
		var maxFreeVRAMGPU library.GPUInfo
		maxFreeVRAM := uint64(0)
		for _, gpu := range allGPUs {
			if gpu.FreeMemory > maxFreeVRAM {
				maxFreeVRAMGPU = gpu
				maxFreeVRAM = gpu.FreeMemory
			}
		}

		if maxFreeVRAMGPU.Vendor == library.NVIDIA {
			chosenGPUVendor = library.NVIDIA
		} else if maxFreeVRAMGPU.Vendor == library.AMD {
			chosenGPUVendor = library.AMD
		} else {
			fmt.Println("Unknown GPU vendor")
			// return here because we need to have at least one GPU
			return models.DeploymentResponse{Success: false, Content: "Unknown GPU vendor."}
		}

		zlog.Sugar().Infoln("GPU with the highest free VRAM on this machine:")
		zlog.Sugar().Infof("Name: %s\n", maxFreeVRAMGPU.GPUName)
		zlog.Sugar().Infof("Total Memory: %d MiB\n", maxFreeVRAMGPU.TotalMemory)
		zlog.Sugar().Infof("Used Memory: %d MiB\n", maxFreeVRAMGPU.UsedMemory)
		zlog.Sugar().Infof("Free Memory: %d MiB\n", maxFreeVRAMGPU.FreeMemory)
		zlog.Sugar().Infof("Chosen GPU Vendor: %v\n", chosenGPUVendor)
	}
	// Pull the image
	imageName := depReq.Params.ImageID
	if chosenGPUVendor == library.AMD {
		imageName += "-amd"
	}
	err := PullImage(ctx, imageName)
	if err != nil {
		zlog.Sugar().Errorf("couldn't pull image: %v", err)
		return models.DeploymentResponse{Success: false, Content: "Unable to pull image."}
	}

	metadata, err := utils.ReadMetadataFile()
	if err != nil {
		zlog.Sugar().Errorf("couldn't read metadata: %v", err)
	}

	// create a service and pass the primary key to the RunContainer to update ContainerID
	var service models.Services
	service.ImageID = imageName
	service.ServiceName = imageName
	service.JobStatus = "running"
	service.JobDuration = 0
	service.EstimatedJobDuration = int64(depReq.Constraints.Time)
	service.TxHash = depReq.TxHash
	service.TransactionType = "running"
	service.MetadataHash = depReq.MetadataHash
	service.WithdrawHash = depReq.WithdrawHash
	service.RefundHash = depReq.RefundHash
	service.Distribute_50Hash = depReq.Distribute_50Hash
	service.Distribute_75Hash = depReq.Distribute_75Hash
	service.ServiceProviderAddr = depReq.RequesterWalletAddress
	service.ComputeProviderAddr = metadata.PublicKey

	// create logbin here and pass it to RunContainer to update logs
	createdLog, err := newLogBin(
		strings.Join(
			[]string{
				depReq.Params.LocalNodeID[:10],
				depReq.Params.RemoteNodeID[:10],
				fmt.Sprintf("%d", time.Now().Unix()),
			},
			"_"))
	if err != nil {
		zlog.Sugar().Errorf("couldn't create log at logbin: %v", err)
		return models.DeploymentResponse{Success: false, Content: "Unable to create log at LogBin."}
	}

	service.LogURL = createdLog.RawUrl
	// Save the service with logs
	if err := db.DB.Create(&service).Error; err != nil {
		zlog.Sugar().Errorf("couldn't save service: %v", err)
		return models.DeploymentResponse{Success: false, Content: "Couldn't save service."}
	}

	// Send service status update
	serviceBytes, _ := json.Marshal(service)
	libp2p.DeploymentUpdate(libp2p.MsgJobStatus, string(serviceBytes), false)

	resCh := make(chan models.DeploymentResponse)

	// Run the container.
	go RunContainer(ctx, depReq, createdLog, resCh, service.ID, chosenGPUVendor)

	res := <-resCh

	// Send back createdLog.RawUrl
	res.Content = createdLog.RawUrl
	return res
}

// createTarballAndChecksum takes containerID and job id and creates a tarball
// of the container along with sha256sum of the tarball.
func createTarballAndChecksum(dc *client.Client, containerID string, callID int64) {
	ctx := context.Background()
	// create a tar of /workspace dir; which is where the image stores it's ML progress
	tarStream, _, err := dc.CopyFromContainer(ctx, containerID, "/workspace")
	if err != nil {
		zlog.Sugar().Fatalf("Error getting tar stream:", err)
	}
	defer tarStream.Close()

	// save the tarball to the designated place on host machine
	checkpointPath := fmt.Sprintf("%s/checkpoints", config.GetConfig().General.MetadataPath)
	// make sure the checkpointPath is an existing directory.
	utils.CreateDirectoryIfNotExists(checkpointPath)

	// use unique "job id" name for the tarball
	tarOnHostPath := filepath.Join(checkpointPath, fmt.Sprintf("%d.tar.gz", callID))
	tarballFile, err := os.Create(tarOnHostPath)
	if err != nil {
		zlog.Sugar().Fatalf("Error creating tarball file:", err)
	}
	defer tarballFile.Close()

	// Create a gzip writer for the tarball file
	gzipWriter := gzip.NewWriter(tarballFile)
	defer gzipWriter.Close()

	// Copy the tar stream to the gzip writer (compressing it)
	_, err = io.Copy(gzipWriter, tarStream)
	if err != nil {
		zlog.Sugar().Fatalf("Error copying tar stream to tarball file:", err)
	}

	zlog.Debug("Tarball of the home directory created successfully.")

	// Calculate the SHA-256 checksum of the tar.gz file
	sha256Checksum, err := utils.CalculateSHA256Checksum(tarOnHostPath)
	if err != nil {
		zlog.Sugar().Fatalf("Error calculating SHA-256 checksum:", err)
	}

	// Write the checksum to a .sha256.txt file
	sha256FilePath := fmt.Sprintf("%s.sha256.txt", tarOnHostPath)
	sha256File, err := os.Create(sha256FilePath)
	if err != nil {
		zlog.Sugar().Fatalf("Error creating SHA-256 checksum file:", err)
	}
	defer sha256File.Close()

	_, err = sha256File.WriteString(sha256Checksum)
	if err != nil {
		zlog.Sugar().Fatalf("Error writing SHA-256 checksum:", err)
	}

	zlog.Debug("SHA-256 checksum written to .sha256.txt file.")
}

func sendBackupToSPD(containerID string, callID int64, spNodeID string) {
	// send the tarball to SPD
	checkpointPath := fmt.Sprintf("%s/checkpoints", config.GetConfig().General.MetadataPath)
	tarOnHostPath := filepath.Join(checkpointPath, fmt.Sprintf("%d.tar.gz", callID))
	spPeerID, err := peer.Decode(spNodeID)
	if err != nil {
		zlog.Sugar().Fatalf("Error decoding spNodeID:", err)
	}

	// Send checksum file
	checksumFile := fmt.Sprintf("%s.sha256.txt", tarOnHostPath)
	libp2p.SendFileToPeer(context.Background(), spPeerID, checksumFile)
	// Send tar file
	libp2p.SendFileToPeer(context.Background(), spPeerID, tarOnHostPath)
}

// beforeContainerTimeout is a event handler which deals with initiating the
// sending of final checkpoint to service provider (sp).
func beforeContainerTimeout(dc *client.Client, containerID string, callID int64, spNodeID string) {
	zlog.Debug("Sending final checkpoint to SPD")
	createTarballAndChecksum(dc, containerID, callID)
	sendBackupToSPD(containerID, callID, spNodeID)
}

// fetchOnboardedResources returns cpuQuota and memoryMax (in bytes) onboarded to nunet
func fetchOnboardedResources() (cpuQuota, memoryMax int64, err error) {
	// call 'nunet info' command internally and get the reserved_cpu, cpu_max and reserved ram
	metadata, err := onboarding.FetchMetadata()
	if err != nil {
		zlog.Sugar().Errorf("Error fetching metadata: %v", err)
		// will return 0, 0, err
		return
	}

	// Proportion=reserved.cpu/resource.cpu_max
	proportion := metadata.Reserved.CPU / metadata.Resource.CPUMax
	// Quota=100000 * Proportion
	cpuQuota = int64(cpuPeriod * proportion)
	memoryMax = metadata.Reserved.Memory * 1024 * 1024 // convert to bytes

	return cpuQuota, memoryMax, nil
}
