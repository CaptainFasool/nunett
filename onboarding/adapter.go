package onboarding

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/statsdb"
)

// InstallRunAdapter takes in metadata and try to run nunet-adapter on current machine.
func InstallRunAdapter(c *gin.Context, hostname string, metadata *models.MetadataV2, cardanoPassive string) {
	var adapterImageTag string
	var deploymentType string
	var tokenomicsAPIName string
	var adapterPrefix string
	if metadata.Network == "nunet-test" {
		adapterImageTag = "test"
		adapterPrefix = "nunet-adapter-test"
		deploymentType = "test"
		tokenomicsAPIName = "testing-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			adapterImageTag = "arm-test"
		}
	} else if metadata.Network == "nunet-staging" {
		adapterImageTag = "staging"
		adapterPrefix = "nunet-adapter-staging"
		deploymentType = "staging"
		tokenomicsAPIName = "staging-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			adapterImageTag = "arm-staging"
		}
	} else if metadata.Network == "nunet-edge" {
		adapterImageTag = "edge"
		adapterPrefix = "nunet-adapter-edge"
		deploymentType = "edge"
		tokenomicsAPIName = "edge-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			adapterImageTag = "arm-edge"
		}
	} else if metadata.Network == "nunet-team" {
		adapterImageTag = "team"
		adapterPrefix = "nunet-adapter-team"
		deploymentType = "team"
		tokenomicsAPIName = "team-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			adapterImageTag = "arm-team"
		}
	} else {
		log.Println("[DMS Adapter] ERROR No such channel " + metadata.Network)
		log.Println("[DMS Adapter] INFO  Onboarding with nunet-test channel")
		adapterImageTag = "test"
		adapterPrefix = "nunet-adapter-test"
		deploymentType = "test"
		tokenomicsAPIName = "testing-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			adapterImageTag = "arm-test"
		}
	}

	adapterImage := "registry.gitlab.com/nunet/nunet-adapter"
	adapterName := adapterPrefix + "-" + hostname

	// truncate adapter name to less than 60 characters for issue #56
	if len(adapterName) > 60 {
		adapterName = adapterName[:60]
	}

	// XXX might be best to generalize gpu/docker.go and use functions from there
	//     implementing afresh here because of gpu specifity there and no networking
	// XXX would also be good to log adapter's output with a rotating log
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	log.Println("[DMS Adapter] INFO  Start pulling nunet-adapter image")
	reader, err := cli.ImagePull(c, fmt.Sprintf("%s:%s", adapterImage, adapterImageTag), types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(ioutil.Discard, reader)
	log.Println("[DMS Adapter] INFO  Done pulling nunet-adapter image")

	envVars := []string{fmt.Sprintf("tokenomics_api_name=%s", tokenomicsAPIName),
		fmt.Sprintf("deployment_type=%s", deploymentType)}

	contConfig := container.Config{
		Image:    fmt.Sprintf("%s:%s", adapterImage, adapterImageTag),
		Env:      envVars,
		Hostname: adapterName,
		Cmd:      []string{"python", "nunet_adapter.py", "60777"},
	}

	hostConfig := container.HostConfig{
		NetworkMode:   container.NetworkMode("host"),
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}

	existingContainers, err := cli.ContainerList(c, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, container := range existingContainers {
		if strings.Contains(container.Names[0], adapterName) &&
			container.Command == "python nunet_adapter.py 60777" &&
			container.HostConfig.NetworkMode == "host" {
			if strings.EqualFold(container.State, "running") {
				log.Println("[DMS Adapter] ERROR There seems to be an adapter container with name: " +
					adapterName + " with ID " + container.ID + " already running. Stopping existing container.")
				err := cli.ContainerStop(c, container.ID, nil)
				if err != nil {
					panic(err)
				}
			} else {
				log.Println("[DMS Adapter] ERROR There seems to be an adapter container with name: " +
					adapterName + " with ID " + container.ID +
					" that is not running (" + container.State + ")" +
					" Removing container.")
			}
			err := cli.ContainerRemove(c, container.ID, types.ContainerRemoveOptions{RemoveVolumes: true, Force: true})
			if err != nil {
				panic(err)
			}
			break
		}
	}
	log.Println("[DMS Adapter] INFO  Creating NuNet Adapter container")
	resp, err := cli.ContainerCreate(c, &contConfig, &hostConfig, nil, nil, adapterName)

	if err != nil {
		panic(err)
	}
	log.Println("[DMS Adapter] INFO  Created NuNet Adapter container")
	log.Println("[DMS Adapter] INFO  Starting NuNet Adapter container")
	if err := cli.ContainerStart(c, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	telemetry.CalcFreeResources()
	go func() {
		for {
			adapter.UpdateAvailableResoruces()
			time.Sleep(time.Second * 10)
		}
	}()
	adapter.UpdateMachinesTable()

	if len(metadata.NodeID) == 0 {
		metadata.NodeID = adapter.GetPeerID()

		// Declare variable for sending requested data on NewDeviceOnboarded function of stats_db
		NewDeviceOnboardParams := models.NewDeviceOnboarded{
			PeerID:        metadata.NodeID,
			CPU:           float32(metadata.Reserved.CPU),
			RAM:           float32(metadata.Reserved.Memory),
			Network:       0.0,
			DedicatedTime: 0.0,
			Timestamp:     float32(statsdb.GetTimestamp()),
		}
		statsdb.NewDeviceOnboarded(NewDeviceOnboardParams)
	}
	// XXX Disabled because of https://gitlab.com/nunet/device-management-service/-/issues/116
	// else {
	// 	statsdb.DeviceResourceChange(metadata)
	// }

	go statsdb.HeartBeat(metadata.NodeID)

	file, _ := json.MarshalIndent(metadata, "", " ")
	err = os.WriteFile("/etc/nunet/metadataV2.json", file, 0644)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "could not write metadata.json"})
		return
	}
}
