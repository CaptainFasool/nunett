package onboarding

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/models"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func InstallRunAdapter(c *gin.Context, hostname string, metadata *models.MetadataV2, cardanoPassive string) {
	var adapterImageTag string
	var deploymentType string
	var tokenomicsApiName string
	var adapterPrefix string
	if metadata.Network == "nunet-development" {
		adapterImageTag = "test"
		adapterPrefix = "testing-nunet-adapter"
		deploymentType = "test"
		tokenomicsApiName = "testing-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			adapterImageTag = "arm-test"
		}
	}
	if metadata.Network == "nunet-private-alpha" {
		adapterImageTag = "latest"
		adapterPrefix = "nunet-adapter"
		deploymentType = "prod"
		tokenomicsApiName = "tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			adapterImageTag = "arm-latest"
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

	log.Println("[DMS Adapter] Start pulling nunet-adapter image")
	reader, err := cli.ImagePull(c, fmt.Sprintf("%s:%s", adapterImage, adapterImageTag), types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(ioutil.Discard, reader)
	log.Println("[DMS Adapter] Done pulling nunet-adapter image")

	envVars := []string{fmt.Sprintf("tokenomics_api_name=%s", tokenomicsApiName),
		fmt.Sprintf("deployment_type=%s", deploymentType)}

	contConfig := container.Config{
		Image:    fmt.Sprintf("%s:%s", adapterImage, adapterImageTag),
		Env:      envVars,
		Hostname: adapterName,
		Cmd:      []string{"python", "nunet_adapter.py", "60777"},
	}

	hostConfig := container.HostConfig{
		NetworkMode: container.NetworkMode("host"),
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
				log.Println("[DMS Adapter] ERROR : There seems to be an adapter container with name: " +
					adapterName + " with ID " +  container.ID + " already running. Stopping existing container.")
				err := cli.ContainerStop(c, container.ID, nil)
				if err != nil {
					panic(err)
				}
			} else {
				log.Println("[DMS Adapter] ERROR : There seems to be an adapter container with name: " +
					adapterName + " with ID " +  container.ID +
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
	log.Println("[DMS Adapter] INFO Creating NuNet Adapter container")
	resp, err := cli.ContainerCreate(c, &contConfig, &hostConfig, nil, nil, adapterName)

	if err != nil {
		panic(err)
	}
	log.Println("[DMS Adapter] INFO : Created NuNet Adapter container")
	log.Println("[DMS Adapter] INFO : Starting NuNet Adapter container")
	if err := cli.ContainerStart(c, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(c, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
}
