package docker

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/google/go-github/github"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/models"
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
	gpuOpts := opts.GpuOpts{}
	gpuOpts.Set("all")

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

	depRes := models.DeploymentResponse{Success: true}
	resCh <- depRes

	tick := time.NewTicker(gistUpdateDuration)
	defer tick.Stop()

	statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	for {
		select {
		case err := <-errCh:
			// handle error & exit
			if err != nil {
				panic(err)
			}
			freeUsedResources(resp.ID)
			return
		case <-statusCh: // not running?
			// get the last logs & exit...
			updateGist(*createdGist.ID, resp.ID)
			freeUsedResources(resp.ID)
			return
		case <-tick.C:
			// get the latest logs ...
			log.Println("updating gist")
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
