package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"

	"runtime"
	"strings"
	"time"

	"github.com/KyleBanks/dockerstats"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/google/go-github/github"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding/gpuinfo"
	"gitlab.com/nunet/device-management-service/plugins/ipfs_plugin"
	"gitlab.com/nunet/device-management-service/utils"
	"go.uber.org/zap"
)

var (
	vcpuToMicroseconds float64       = 100000
	gistUpdateInterval time.Duration = time.Duration(config.GetConfig().Job.GistUpdateInterval) * time.Minute
)

func freeUsedResources() {
	// update the available resources table
	err := telemetry.CalcFreeResources()
	if err != nil {
		zlog.Sugar().Errorf("Error getting freeResources: %v", err)
	}
	libp2p.UpdateKadDHT()
}

func mhzPerCore() (float64, error) {
	cpus, err := cpu.Info()
	if err != nil {
		zlog.Sugar().Errorf("failed to get cpu info: %v", err)
		return 0, err
	}
	return cpus[0].Mhz, nil
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

// RunContainer goes through the process of setting constraints,
// specifying image name and cmd. It starts a container and posts
// log update every gistUpdateDuration.
func RunContainer(depReq models.DeploymentRequest, createdGist *github.Gist, resCh chan<- models.DeploymentResponse, servicePK uint, chosenGPUVendor gpuinfo.GPUVendor) {
	zlog.Info("Entering RunContainer")
	machine_type := depReq.Params.MachineType
	gpuOpts := opts.GpuOpts{}
	if machine_type == "gpu" {
		gpuOpts.Set("all") // TODO find a way to use GPU and CPU
	}
	modelURL := depReq.Params.ModelURL
	packages := strings.Join(depReq.Params.Packages, " ")
	containerConfig := &container.Config{
		Image: depReq.Params.ImageID,
		Cmd:   []string{modelURL, packages},
		// Tty:          true,
	}

	memoryMbToBytes := int64(depReq.Constraints.RAM * 1024 * 1024)
	VCPU, err := mhzToVCPU(depReq.Constraints.CPU)
	if err != nil {
		zlog.Sugar().Errorf("Error converting MHz to VCPU: %v", err)
		depRes := models.DeploymentResponse{Success: false, Content: "Problem with CPU Constraints. Unable to process request."}
		resCh <- depRes
		return
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			DeviceRequests: gpuOpts.Value(),
			Memory:         memoryMbToBytes,
			CPUQuota:       int64(VCPU * vcpuToMicroseconds),
		},
	}

	hostConfigAMDGPU := container.HostConfig{}

	if chosenGPUVendor == gpuinfo.AMD {
		hostConfigAMDGPU = container.HostConfig{
			Binds: []string{
				"/dev/kfd:/dev/kfd",
				"/dev/dri:/dev/dri",
			},
			Resources: container.Resources{
				Memory:   memoryMbToBytes,
				CPUQuota: int64(VCPU * vcpuToMicroseconds),
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

		if runtime.GOOS != "ubuntu18.04" {
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
	// TODO: Update service based on passed pk

	telemetry.CalcFreeResources()
	libp2p.UpdateKadDHT()

	depRes := models.DeploymentResponse{Success: true}
	resCh <- depRes

	tick := time.NewTicker(gistUpdateInterval)
	defer tick.Stop()

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
			updateGist(*createdGist.ID, resp.ID)

			// Add a response for log update
			if r := db.DB.Where("container_id = ?", resp.ID).First(&service); r.Error != nil {
				zlog.Sugar().Errorf("problem updating services: %v", r.Error)
				service.JobStatus = "finished with errors"
			}
			sendLogsToSPD(resp.ID, service.LastLogFetch.Format("2006-01-02T15:04:05Z"))

			var requestTracker models.RequestTracker
			res := db.DB.Where("id = ?", 1).Find(&requestTracker)
			if res.Error != nil {
				zlog.Error(res.Error.Error())
			}

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

			db.DB.Save(&service)

			// Send service status update
			serviceBytes, _ := json.Marshal(service)
			var closeStream bool
			if strings.Contains(string(service.JobStatus), "finished") {
				closeStream = true
			}
			libp2p.DeploymentUpdate(libp2p.MsgJobStatus, string(serviceBytes), closeStream)

			res = db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Updates(requestTracker)
			if res.Error != nil {
				zlog.Sugar().Errorf("problem updating request tracker: %v", res.Error)
				depRes := models.DeploymentResponse{Success: false, Content: "Problem with request tracker. Unable to complete request."}
				resCh <- depRes
				return
			}

			if utils.SliceContainsValue("outputIPFS", depReq.Params.AdditionalFeatures) {
				zlog.Sugar().Info("SP chose to store output on IPFS, calling plugin")
				go ipfs_plugin.UseOutputFeatIPFS(resp.ID)
			}

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

			updateGist(*createdGist.ID, resp.ID)

			// Add a response for log update
			db.DB.Where("container_id = ?", resp.ID).First(&service)
			zlog.Debug("service.LastLogFetch",
				zap.String("value", service.LastLogFetch.Format("2006-01-02T15:04:05Z")),
			)
			sendLogsToSPD(resp.ID, service.LastLogFetch.Format("2006-01-02T15:04:05Z"))
			service.LastLogFetch = time.Now().In(time.UTC)
			db.DB.Save(&service)
		}
	}
}

// PullImage is a wrapper around Docker SDK's function with same name.
func PullImage(imageName string) error {
	out, err := dc.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("unable to pull image: %v", err)
	}

	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
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

// HandleDeployment does following docker based actions in the sequence:
// Pull image, run container, get logs, delete container, send log to the requester
func HandleDeployment(depReq models.DeploymentRequest) models.DeploymentResponse {
	var chosenGPUVendor gpuinfo.GPUVendor
	if depReq.Params.MachineType == "gpu" {
		// Finding the GPU with the highest free VRAM regardless of vendor type
		// Get AMD GPU info
		amdGPUs, err := gpuinfo.GetAMDGPUInfo()
		if err != nil {
			zlog.Sugar().Errorf("AMD GPU/Driver not found: %v", err)
		}

		// Get NVIDIA GPU info
		nvidiaGPUs, err := gpuinfo.GetNVIDIAGPUInfo()
		if err != nil {
			zlog.Sugar().Errorf("NVIDIA GPU/Driver not found: %v", err)
			// return here and not above for AMD because we need to have at least one GPU
			return models.DeploymentResponse{Success: false, Content: "Unable to get GPU info."}
		}

		// Combine AMD and NVIDIA GPU info
		allGPUs := append(amdGPUs, nvidiaGPUs...)

		// Find the GPU with the highest free VRAM
		var maxFreeVRAMGPU gpuinfo.GPUInfo
		maxFreeVRAM := uint64(0)
		for _, gpu := range allGPUs {
			if gpu.FreeMemory > maxFreeVRAM {
				maxFreeVRAMGPU = gpu
				maxFreeVRAM = gpu.FreeMemory
			}
		}

		if maxFreeVRAMGPU.Vendor == gpuinfo.NVIDIA {
			chosenGPUVendor = gpuinfo.NVIDIA
		} else if maxFreeVRAMGPU.Vendor == gpuinfo.AMD {
			chosenGPUVendor = gpuinfo.AMD
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
	if chosenGPUVendor == gpuinfo.AMD {
		imageName += "-amd"
	}
	err := PullImage(imageName)
	if err != nil {
		zlog.Sugar().Errorf("couldn't pull image: %v", err)
		return models.DeploymentResponse{Success: false, Content: "Unable to pull image."}
	}

	// create a service and pass the primary key to the RunContainer to update ContainerID
	var service models.Services
	service.ImageID = depReq.Params.ImageID
	service.ServiceName = depReq.Params.ImageID
	service.JobStatus = "running"
	service.JobDuration = 5           // these are dummy data, implementation pending
	service.EstimatedJobDuration = 10 // these are dummy data, implementation pending

	// create gist here and pass it to RunContainer to update logs
	createdGist, _, err := createGist()
	if err != nil {
		zlog.Sugar().Errorf("couldn't create gist: %v", err)
		return models.DeploymentResponse{Success: false, Content: "Unable to create Gist."}
	}

	service.LogURL = *createdGist.HTMLURL
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
	go RunContainer(depReq, createdGist, resCh, service.ID, chosenGPUVendor)

	res := <-resCh

	// Send back *createdGist.HTMLURL
	res.Content = *createdGist.HTMLURL
	return res
}
