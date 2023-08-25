package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/models"
)

var (
	vcpuToMicroseconds float64 = 100000
)

type PreStartCallback func()
type RunningCallback func()
type FailedCallback func()
type SucceededCallback func()

type DockerJob struct {
	log          *logger.Logger
	ctx          *context.Context
	depReq       *models.DeploymentRequest
	tickInterval time.Duration
	spdtickInterval time.Duration
	dc           *client.Client
}

// NewJob creates a new job, let it be docker conatiner or something else.
func NewDockerJob(depReq *models.DeploymentRequest) *DockerJob {
	return &DockerJob{
		depReq:       depReq,
		tickInterval: time.Duration(config.GetConfig().Job.LogUpdateInterval) * time.Minute,
		spdtickInterval: 10 * time.Second
	}

}

func (dj *DockerJob) Run(
	beforeStart PreStartCallback,
	running RunningCallback,
	failed FailedCallback,
	succeeded SucceededCallback,
) error {

	dj.log.Info("Entering DockerJob.Run()")
	machineType := dj.depReq.Params.MachineType
	gpuOpts := opts.GpuOpts{}
	if machineType == "gpu" {
		gpuOpts.Set("all") // TODO find a way to use GPU and CPU
	}
	modelURL := dj.depReq.Params.ModelURL
	packages := strings.Join(dj.depReq.Params.Packages, " ")
	containerConfig := &container.Config{
		Image: dj.depReq.Params.ImageID,
		Cmd:   []string{modelURL, packages},
		// Tty:          true,
	}
	memoryMbToBytes := int64(dj.depReq.Constraints.RAM * 1024 * 1024)
	VCPU := mhzToVCPU(dj.depReq.Constraints.CPU)

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			DeviceRequests: gpuOpts.Value(),
			Memory:         memoryMbToBytes,
			CPUQuota:       int64(VCPU * vcpuToMicroseconds),
		},
	}

	beforeStart()

	// TODO 1 start: This needs to be handled outside
	// var freeRes models.FreeResources

	// if res := db.DB.Find(&freeRes); res.RowsAffected == 0 {
	// 	dj.log.Sugar().Errorf("Record not found!")
	// 	depRes := models.DeploymentResponse{Success: false, Content: "Problem with Free Resources. Unable to process request."}
	// 	resCh <- depRes
	// 	return
	// }

	// TODO 1 ends

	// TODO 2 start
	// Check if we have enough free resources before running Container
	// if (depReq.Constraints.RAM > freeRes.Ram) ||
	// 	(depReq.Constraints.CPU > freeRes.TotCpuHz) {
	// 	dj.log.Sugar().Errorf("Not enough resources available to deploy container")
	// 	depRes := models.DeploymentResponse{Success: false, Content: "Problem with resources for deployment. Unable to process request."}
	// 	resCh <- depRes
	// 	return
	// }
	// TODO 2 ends

	resp, err := dj.dc.ContainerCreate(*dj.ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}

	if err := dj.dc.ContainerStart(*dj.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// TODO 3 start
	// var requestTracker models.RequestTracker
	// res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	// if res.Error != nil {
	// 	dj.log.Error(res.Error.Error())
	// }
	// status := "started"

	// ServiceStatusParams := models.ServiceStatus{
	// 	CallID:              requestTracker.CallID,
	// 	PeerIDOfServiceHost: requestTracker.NodeID,
	// 	Status:              status,
	// 	Timestamp:           float32(statsdb.GetTimestamp()),
	// }
	// statsdb.ServiceStatus(ServiceStatusParams)

	// updating RequestTracker
	// requestTracker.Status = status
	// requestTracker.RequestID = resp.ID
	// res = db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Updates(requestTracker)
	// if res.Error != nil {
	// 	dj.log.Sugar().Errorf("unable to update request tracker: %v", res.Error)
	// 	depRes := models.DeploymentResponse{Success: false, Content: "Problem with request tracker. Unable to process request."}
	// 	resCh <- depRes
	// 	return
	// }
	// TODO 3 ends

	// // Update db - find the service based on primary key and update container id
	// var service models.Services
	// res = db.DB.Model(&service).Where("id = ?", servicePK).Updates(models.Services{ContainerID: resp.ID})
	// if res.Error != nil {
	// 	dj.log.Sugar().Errorf("unable to update services: %v", res.Error)
	// 	depRes := models.DeploymentResponse{Success: false, Content: "Problem with services tracker. Unable to process request."}
	// 	resCh <- depRes
	// 	return
	// }

	// var resourceRequirements models.ServiceResourceRequirements
	// resourceRequirements.CPU = depReq.Constraints.CPU
	// resourceRequirements.RAM = depReq.Constraints.RAM

	// result := db.DB.Create(&resourceRequirements)
	// if result.Error != nil {
	// 	dj.log.Sugar().Errorf("unable to create resource requirements: %v", res.Error)
	// 	depRes := models.DeploymentResponse{Success: false, Content: "Problem with resource requirements. Unable to process request."}
	// 	resCh <- depRes
	// 	return
	// }

	// service.ResourceRequirements = int(resourceRequirements.ID)

	// telemetry.CalcFreeResources()
	// freeResource, err := telemetry.GetFreeResources()
	// if err != nil {
	// 	dj.log.Sugar().Errorf("Error getting freeResources: %v", err)
	// } else {
	// 	statsdb.DeviceResourceChange(freeResource)
	// }

	// libp2p.UpdateDHT()

	// depRes := models.DeploymentResponse{Success: true}
	// resCh <- depRes

	tick := time.NewTicker(dj.tickInterval)
	defer tick.Stop()

	spdtick := time.NewTicker(spdtickInterval)
	defer spdTick.Stop()

	statusCh, errCh := dj.dc.ContainerWait(*dj.ctx, resp.ID, container.WaitConditionNotRunning)
	// maxUsedRAM, maxUsedCPU, networkBwUsed := 0.0, 0.0, 0.0

outerLoop:
	for {
		select {
		case err := <-errCh:
			dj.log.Info("[container running] entering first case; container didn't start")
			_ = err
			failed()

			// handle error & exit
			// if err != nil {
			// What to do here?
			// dj.log.Sugar().Errorf("problem in running contianer: %v", err)
			// depRes := models.DeploymentResponse{Success: false, Content: "Problem occurred with container. Unable to complete request."}
			// resCh <- depRes
			// freeUsedResources(resp.ID)
			// return
			// }
		case containerStatus := <-statusCh: // not running?
			dj.log.Info("[container running] entering second case; container exiting")
			_ = containerStatus
			// get the last logs & exit...
			// updateGist(*createdGist.ID, resp.ID)

			// // Add a response for log update
			// if r := db.DB.Where("container_id = ?", resp.ID).First(&service); r.Error != nil {
			// 	dj.log.Sugar().Errorf("problemn updating services: %v", r.Error)
			// 	service.JobStatus = "finished with errors"
			// }
			// sendLogsToSPD(resp.ID, service.LastLogFetch.Format("2006-01-02T15:04:05Z"))

			// var requestTracker models.RequestTracker
			// res := db.DB.Where("id = ?", 1).Find(&requestTracker)
			// if res.Error != nil {
			// 	dj.log.Error(res.Error.Error())
			// }

			// // add exitStatus to db
			// if containerStatus.StatusCode == 0 {
			// 	service.JobStatus = libp2p.ContainerJobFinishedWithoutErrors
			// 	status = libp2p.ContainerJobFinishedWithoutErrors
			// 	requestTracker.Status = libp2p.ContainerJobFinishedWithoutErrors
			// } else if containerStatus.StatusCode > 0 {
			// 	service.JobStatus = libp2p.ContainerJobFinishedWithErrors
			// 	status = libp2p.ContainerJobFinishedWithErrors
			// 	requestTracker.Status = libp2p.ContainerJobFinishedWithErrors
			// }

			// TODO: Send running or failed based on containerStatus.StatusCode above

			// db.DB.Save(&service)

			// Send service status update
			// serviceBytes, _ := json.Marshal(service)
			// var closeStream bool
			// if strings.HasPrefix("finished", service.JobStatus) {
			// 	closeStream = true
			// }
			// libp2p.DeploymentUpdate(libp2p.MsgJobStatus, string(serviceBytes), closeStream)

			// duration := time.Now().Sub(service.CreatedAt)
			// timeTaken := duration.Seconds()
			// averageNetBw := networkBwUsed / timeTaken
			// ServiceCallParams := models.ServiceCall{
			// 	CallID:              requestTracker.CallID,
			// 	PeerIDOfServiceHost: depReq.Params.LocalNodeID,
			// 	CPUUsed:             float32(maxUsedCPU),
			// 	MemoryUsed:          float32(maxUsedRAM),
			// 	MaxRAM:              float32(depReq.Constraints.RAM),
			// 	NetworkBwUsed:       float32(averageNetBw),
			// 	TimeTaken:           float32(timeTaken),
			// 	Status:              status,
			// 	AmountOfNtx:         requestTracker.MaxTokens,
			// 	Timestamp:           float32(statsdb.GetTimestamp()),
			// }
			// statsdb.ServiceCall(ServiceCallParams)

			// res = db.DB.Model(&models.RequestTracker{}).Where("id = ?", 1).Updates(requestTracker)
			// if res.Error != nil {
			// 	dj.log.Sugar().Errorf("problem updating request tracker: %v", res.Error)
			// 	depRes := models.DeploymentResponse{Success: false, Content: "Problem with request tracker. Unable to complete request."}
			// 	resCh <- depRes
			// 	return
			// }
			// freeUsedResources(resp.ID)
			break outerLoop
		case <-tick.C:
			dj.log.Info("[container running] entering third case; time ticker")

			running()

			// // get the latest logs ...
			// // contID := requestTracker.RequestID[:12]
			// stats, err := dockerstats.Current()
			// if err != nil {
			// 	dj.log.Sugar().Errorf("problem obtaining docker stats: %v", err)
			// }

			// var tempService models.Services
			// if err := db.DB.Where("container_id = ?", resp.ID).First(&tempService).Error; err != nil {
			// 	panic(err)
			// }
			// duration := time.Since(tempService.CreatedAt)
			// timeTaken := duration.Seconds()
			// for _, s := range stats {
			// 	if s.Container == contID {
			// 		usedRAM := strings.Split(s.Memory.Raw, "/")
			// 		usedCPU := strings.Split(s.CPU, "%")
			// 		usedNetwork := strings.Split(s.IO.Network, "/")
			// 		ramFloat64 := calculateResourceUsage(usedRAM[0])
			// 		cpuFloat64, _ := strconv.ParseFloat(usedCPU[0], 64)
			// 		cpuFloat64 = cpuUsage(cpuFloat64, float64(depReq.Constraints.CPU))
			// 		networkFloat64Pre := calculateResourceUsage(usedNetwork[0])
			// 		networkFloat64Suf := calculateResourceUsage(usedNetwork[1])
			// 		networkFloat64 := networkFloat64Pre + networkFloat64Suf
			// 		if ramFloat64 > maxUsedRAM {
			// 			maxUsedRAM = ramFloat64 / 1024
			// 		}
			// 		if cpuFloat64 > maxUsedCPU {
			// 			maxUsedCPU = cpuFloat64
			// 		}
			// 		if networkFloat64 > networkBwUsed {
			// 			networkBwUsed = networkFloat64
			// 		}
			// 	}
			// }
			// averageNetBw := networkBwUsed / timeTaken
			// ServiceCallParams := models.ServiceCall{
			// 	CallID:              requestTracker.CallID,
			// 	PeerIDOfServiceHost: depReq.Params.LocalNodeID,
			// 	CPUUsed:             float32(maxUsedCPU),
			// 	MemoryUsed:          float32(maxUsedRAM),
			// 	MaxRAM:              float32(depReq.Constraints.RAM),
			// 	NetworkBwUsed:       float32(averageNetBw),
			// 	TimeTaken:           0.0,
			// 	Status:              "started",
			// 	AmountOfNtx:         requestTracker.MaxTokens,
			// 	Timestamp:           float32(statsdb.GetTimestamp()),
			// }
			// statsdb.ServiceCall(ServiceCallParams)

			// updateGist(*createdGist.ID, resp.ID)

			// // Add a response for log update
			// db.DB.Where("container_id = ?", resp.ID).First(&service)
			// dj.log.Debug("service.LastLogFetch",
			// 	zap.String("value", service.LastLogFetch.Format("2006-01-02T15:04:05Z")),
			// )
			// sendLogsToSPD(resp.ID, service.LastLogFetch.Format("2006-01-02T15:04:05Z"))
			// service.LastLogFetch = time.Now().In(time.UTC)
			// db.DB.Save(&service)
		case <-spdtick.C:
			dj.log.Info("[spd container running] entering fourth case; spd console")

			running()
		}
	}
	return nil
}

func mhzPerCore() (float64, error) {
	cpus, err := cpu.Info()
	if err != nil {
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

func mhzToVCPU(cpuInMhz int) float64 {
	mhz, err := mhzPerCore()
	if err != nil {
		return 0
	}
	vcpu := float64(cpuInMhz) / mhz
	return toFixed(vcpu, 2)
}
