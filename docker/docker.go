package docker

import (
	"bufio"
	"bytes"
	"io"
	"math"
	"os"
	"strconv"

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
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/statsdb"
)

const (
	vcpuToMicroseconds float64       = 100000
	gistUpdateDuration time.Duration = 2 * time.Minute
)

func freeUsedResources(contID string) {
	// update the available resources table
	telemetry.CalcFreeResources()
	libp2p.UpdateDHT()
}

func mhzPerCore() float64 {
	cpus, err := cpu.Info()
	if err != nil {
		panic(err)
	}
	return cpus[0].Mhz
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func mhzToVCPU(cpuInMhz int) float64 {
	vcpu := float64(cpuInMhz) / mhzPerCore()
	return toFixed(vcpu, 2)
}

// RunContainer goes through the process of setting constraints,
// specifying image name and cmd. It starts a container and posts
// log update every gistUpdateDuration.
func RunContainer(depReq models.DeploymentRequest, createdGist *github.Gist, resCh chan<- models.DeploymentResponse, servicePK uint) {
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
	VCPU := mhzToVCPU(depReq.Constraints.CPU)
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			DeviceRequests: gpuOpts.Value(),
			Memory:         memoryMbToBytes,
			CPUQuota:       int64(VCPU * vcpuToMicroseconds),
		},
	}

	var freeRes models.FreeResources

	if res := db.DB.Find(&freeRes); res.RowsAffected == 0 {
		panic("Record not found!")
	}

	// Check if we have enough free resources before running Container
	if (depReq.Constraints.RAM > freeRes.Ram) ||
		(depReq.Constraints.CPU > freeRes.TotCpuHz) {
		panic("Not enough resources available to deploy container")

	}

	resp, err := dc.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")

	if err != nil {
		panic(err)
	}

	if err := dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	var requestTracker models.RequestTracker
	res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	}
	status := "started"

	ServiceStatusParams := models.ServiceStatus{
		CallID:              requestTracker.CallID,
		PeerIDOfServiceHost: requestTracker.NodeID,
		ServiceID:           requestTracker.ServiceType,
		Status:              status,
		Timestamp:           float32(statsdb.GetTimestamp()),
	}
	statsdb.ServiceStatus(ServiceStatusParams)

	// updating RequestTracker
	requestTracker.Status = status
	requestTracker.RequestID = resp.ID
	res = db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Updates(requestTracker)
	if res.Error != nil {
		panic(res.Error)
	}

	// Update db - find the service based on primary key and update container id
	var service models.Services
	res = db.DB.Model(&service).Where("id = ?", servicePK).Updates(models.Services{ContainerID: resp.ID})
	if res.Error != nil {
		panic(res.Error)
	}

	var resourceRequirements models.ServiceResourceRequirements
	resourceRequirements.CPU = depReq.Constraints.CPU
	resourceRequirements.RAM = depReq.Constraints.RAM

	result := db.DB.Create(&resourceRequirements)
	if result.Error != nil {
		panic(result.Error)
	}

	service.ResourceRequirements = int(resourceRequirements.ID)
	// TODO: Update service based on passed pk

	telemetry.CalcFreeResources()
	libp2p.UpdateDHT()

	depRes := models.DeploymentResponse{Success: true}
	resCh <- depRes

	tick := time.NewTicker(gistUpdateDuration)
	defer tick.Stop()

	statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	maxUsedRAM, maxUsedCPU := 0.0, 0.0

outerLoop:
	for {
		select {
		case err := <-errCh:
			// handle error & exit
			if err != nil {
				panic(err)
			}
			freeUsedResources(resp.ID)
			return
		case containerStatus := <-statusCh: // not running?
			// get the last logs & exit...
			updateGist(*createdGist.ID, resp.ID)
			var requestTracker models.RequestTracker
			res := db.DB.Where("id = ?", 1).Find(&requestTracker)
			if res.Error != nil {
				zlog.Error(res.Error.Error())
			}

			// add exitStatus to db
			var services models.Services
			if containerStatus.StatusCode == 0 {
				services.JobStatus = "finished without errors"
				ServiceCallParams := models.ServiceCall{
					CallID:              requestTracker.CallID,
					PeerIDOfServiceHost: requestTracker.NodeID,
					ServiceID:           requestTracker.ServiceType,
					CPUUsed:             float32(maxUsedCPU),
					MaxRAM:              float32(depReq.Constraints.Vram),
					MemoryUsed:          float32(maxUsedRAM),
					NetworkBwUsed:       0.0,
					TimeTaken:           0.0,
					Status:              "finished without errors",
					Timestamp:           float32(statsdb.GetTimestamp()),
				}
				statsdb.ServiceCall(ServiceCallParams)
				requestTracker.Status = "finished without errors"
			} else if containerStatus.StatusCode > 0 {
				services.JobStatus = "finished with errors"
				ServiceStatusParams.CallID = requestTracker.CallID
				ServiceStatusParams.PeerIDOfServiceHost = requestTracker.NodeID
				ServiceStatusParams.Status = "finished with errors"
				ServiceStatusParams.Timestamp = float32(statsdb.GetTimestamp())

				statsdb.ServiceStatus(ServiceStatusParams)
				requestTracker.Status = "finished with errors"
			}
			r := db.DB.Model(services).Where("container_id = ?", resp.ID).Updates(services)
			if r.Error != nil {
				panic(r.Error)
			}

			res = db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Updates(requestTracker)
			if res.Error != nil {
				panic(res.Error)
			}
			freeUsedResources(resp.ID)
			break outerLoop
		case <-tick.C:
			// get the latest logs ...
			zlog.Info("updating gist")
			contID := requestTracker.RequestID[:12]
			stats, err := dockerstats.Current()
			if err != nil {
				panic(err)
			}

			for _, s := range stats {
				if s.Container == contID {
					usedRAM := strings.Split(s.Memory.Raw, "MiB")
					usedCPU := strings.Split(s.CPU, "%")
					ramFloat64, _ := strconv.ParseFloat(usedRAM[0], 64)
					cpuFloat64, _ := strconv.ParseFloat(usedCPU[0], 64)
					cpuFloat64 = cpuUsage(cpuFloat64, float64(hostConfig.CPUQuota))
					if ramFloat64 > maxUsedRAM {
						maxUsedRAM = ramFloat64
					}
					if cpuFloat64 > maxUsedCPU {
						maxUsedCPU = cpuFloat64
					}
				}
			}

			updateGist(*createdGist.ID, resp.ID)
		}
	}

}

// cleanFlushInfo takes in bytes.Buffer from docker logs output and for each line
// if it has a \r in the lines, takes the last one and compose another string
// out of that.
func cleanFlushInfo(bytesBuffer *bytes.Buffer) string {
	scanner := bufio.NewScanner(bytesBuffer)
	finalString := ""

	for scanner.Scan() {
		line := scanner.Text()
		chunks := strings.Split(line, "\r")
		lastChunk := chunks[len(chunks)-1] // fetch the last update of the line
		finalString += lastChunk + "\n"
	}

	return finalString
}

// PullImage is a wrapper around Docker SDK's function with same name.
func PullImage(imageName string) {
	out, err := dc.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer out.Close()
	io.Copy(os.Stdout, out)
}

// GetLogs return logs from the container io.ReadCloser. It's the caller duty
// duty to do a stdcopy.StdCopy. Any other method might render unknown
// unicode character as log output has both stdout and stderr. That starting
// has info if that line is stderr or stdout.
func GetLogs(contName string) (logOutput io.ReadCloser) {
	options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}

	out, err := dc.ContainerLogs(ctx, contName, options)
	if err != nil {
		panic(err)
	}

	return out
}

func cpuUsage(cpu float64, maxCPU float64) float64 {
	return maxCPU * cpu / 100
}

// HandleDeployment does following docker based actions in the sequence:
// Pull image, run container, get logs, delete container, send log to the requester
func HandleDeployment(depReq models.DeploymentRequest) models.DeploymentResponse {
	// Pull the image
	imageName := depReq.Params.ImageID
	PullImage(imageName)

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
		zlog.Sugar().Errorln(err)
	}

	service.LogURL = *createdGist.HTMLURL
	// Save the service with logs
	if err := db.DB.Create(&service).Error; err != nil {
		panic(err)
	}
	resCh := make(chan models.DeploymentResponse)

	// Run the container.
	go RunContainer(depReq, createdGist, resCh, service.ID)

	res := <-resCh

	// Send back *createdGist.HTMLURL
	res.Content = *createdGist.HTMLURL
	return res
}
