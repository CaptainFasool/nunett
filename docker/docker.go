package docker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KyleBanks/dockerstats"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/go-github/github"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	//"gitlab.com/nunet/device-management-service/statsdb" //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
)

const (
	vcpuToMicroseconds float64       = 100000
	gistUpdateDuration time.Duration = 2 * time.Minute
)

func freeUsedResources(contID string) {
	// Remove Service from Services table
	// and update the available resources table
	var service []models.Services
	result := db.DB.Where("container_id = ?", contID).Find(&service)
	if result.Error != nil {
		panic(result.Error)
	}
	db.DB.Delete(&service)

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
func RunContainer(depReq models.DeploymentRequest, createdGist *github.Gist, resCh chan<- models.DeploymentResponse) {
	log.Println("Entering RunContainer")
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

	peerIDOfServiceHost := depReq.Params.NodeID
	status := "started"
	var requestTracker models.RequestTracker

	if res := db.DB.Where("node_id = ?", peerIDOfServiceHost).Find(&requestTracker); res.RowsAffected == 0 {
		panic("Service Not Found for Deployment")
	}
	// ServiceRunParams := models.ServiceRun{  //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
	// 	CallID:              requestTracker.CallID,
	// 	PeerIDOfServiceHost: peerIDOfServiceHost,
	// 	Status:              status,
	// 	Timestamp:           float32(statsdb.GetTimestamp()),
	// }

	//statsdb.ServiceRun(ServiceRunParams) //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
	//update RequestTracker
	requestTracker.Status = status
	requestTracker.RequestID = resp.ID
	res := db.DB.Model(&models.RequestTracker{}).Where("node_id = ?", peerIDOfServiceHost).Updates(requestTracker)
	if res.Error != nil {
		panic(res.Error)
	}

	// Update db
	var service models.Services
	service.ContainerID = resp.ID
	service.ImageID = depReq.Params.ImageID
	service.ServiceName = depReq.Params.ImageID

	var resourceRequirements models.ServiceResourceRequirements
	resourceRequirements.CPU = depReq.Constraints.CPU
	resourceRequirements.RAM = depReq.Constraints.RAM

	result := db.DB.Create(&resourceRequirements)
	if result.Error != nil {
		panic(result.Error)
	}

	service.ResourceRequirements = int(resourceRequirements.ID)
	result = db.DB.Create(&service)
	if result.Error != nil {
		panic(result.Error)
	}

	telemetry.CalcFreeResources()
	libp2p.UpdateDHT()

	depRes := models.DeploymentResponse{Success: true}
	resCh <- depRes

	tick := time.NewTicker(gistUpdateDuration)
	defer tick.Stop()

	statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	maxUsedRAM, maxUsedCPU := 0.0, 0.0
	for {
		select {
		case err := <-errCh:
			// add also exit status to Gist
			exitStatus := bytes.NewBuffer([]byte(""))
			exitErr := bytes.NewBuffer([]byte(err.Error()))
			updateGist(*createdGist.ID, *exitStatus, *exitErr)

			// handle error & exit
			if err != nil {
				panic(err)
			}
			if res := db.DB.Where("request_id= ?", resp.ID).Find(&requestTracker); res.RowsAffected == 0 {
				panic("Service Not Found for Deployment")
			}

			// ServiceRunParams.CallID = requestTracker.CallID  //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
			// ServiceRunParams.PeerIDOfServiceHost = requestTracker.NodeID
			// ServiceRunParams.Status = "unknown"
			// ServiceRunParams.Timestamp = float32(statsdb.GetTimestamp())

			// statsdb.ServiceRun(ServiceRunParams) //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
			requestTracker.Status = "unknown"
			res := db.DB.Model(&models.RequestTracker{}).Where("request_id = ?", resp.ID).Updates(requestTracker)
			if res.Error != nil {
				panic(res.Error)
			}
			freeUsedResources(resp.ID)
			return
		case status := <-statusCh: // not running?
			// get the last logs & exit...
			stdout, stderr := GetLogs(resp.ID)
			updateGist(*createdGist.ID, stdout, stderr)
			// add also exit status to Gist
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(status.StatusCode))
			exitStatus := bytes.NewBuffer(b)
			exitErr := bytes.NewBuffer([]byte(""))
			updateGist(*createdGist.ID, *exitStatus, *exitErr)

			if res := db.DB.Where("request_id= ?", resp.ID).Find(&requestTracker); res.RowsAffected == 0 {
				panic("Service Not Found for Deployment")
			}
			if depRes.Success {
				// ServiceCallParams := models.ServiceCall{   //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
				// 	CallID:              requestTracker.CallID,
				// 	PeerIDOfServiceHost: requestTracker.NodeID,
				// 	ServiceID:           requestTracker.ServiceType,
				// 	CPUUsed:             float32(maxUsedCPU),
				// 	MaxRAM:              float32(depReq.Constraints.Vram),
				// 	MemoryUsed:          float32(maxUsedRAM),
				// 	NetworkBwUsed:       0.0,
				// 	TimeTaken:           0.0,
				// 	Status:              "success",
				// 	Timestamp:           float32(statsdb.GetTimestamp()),
				// }
				// statsdb.ServiceCall(ServiceCallParams)
				requestTracker.Status = "success"
			} else if !depRes.Success {
				// ServiceRunParams.CallID = requestTracker.CallID   //XXX: Disabled StatsDB Calls - Refer to https://gitlab.com/nunet/device-management-service/-/issues/138
				// ServiceRunParams.PeerIDOfServiceHost = requestTracker.NodeID
				// ServiceRunParams.Status = "failed"
				// ServiceRunParams.Timestamp = float32(statsdb.GetTimestamp())

				// statsdb.ServiceRun(ServiceRunParams)
				requestTracker.Status = "failed"
			}

			res := db.DB.Model(&models.RequestTracker{}).Where("request_id = ?", resp.ID).Updates(requestTracker)
			if res.Error != nil {
				panic(res.Error)
			}
			freeUsedResources(resp.ID)
			return
		case <-tick.C:
			// get the latest logs ...
			log.Println("updating gist")

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
			stdout, stderr := GetLogs(resp.ID)
			updateGist(*createdGist.ID, stdout, stderr)
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
func GetLogs(containerID string) (bytes.Buffer, bytes.Buffer) {
	options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}

	out, err := dc.ContainerLogs(ctx, containerID, options)
	if err != nil {
		panic(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	stdcopy.StdCopy(&stdout, &stderr, out)

	return stdout, stderr
}

func cpuUsage(cpu float64, maxCPU float64) float64 {
	return maxCPU * cpu / 100
}

// HandleDeployment does following docker based actions in the sequence:
// Pull image, run container, get logs, delete container, send log to the requester
func HandleDeployment(depReq models.DeploymentRequest, depRes models.DeploymentResponse) models.DeploymentResponse {
	// Pull the image
	imageName := depReq.Params.ImageID
	PullImage(imageName)

	createdGist, _, err := createGist()
	if err != nil {
		log.Panicln(err)
	}

	resCh := make(chan models.DeploymentResponse)

	// Run the container.
	go RunContainer(depReq, createdGist, resCh)

	res := <-resCh

	// Send back *createdGist.HTMLURL
	depRes.Content = *createdGist.HTMLURL
	depRes.Success = res.Success
	return depRes
}
