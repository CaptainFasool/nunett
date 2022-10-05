package onboarding

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

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
		Image:       fmt.Sprintf("%s:%s", adapterImage, adapterImageTag),
		Env:         envVars,
		Hostname:    adapterName,
		Cmd:         []string{"python", "nunet_adapter.py", "60777"},
	}

	hostConfig := container.HostConfig{}
	hostConfig.NetworkMode = container.NetworkMode("host")

	log.Println("[DMS Adapter] Creating NuNet Adapter container")
	resp, err := cli.ContainerCreate(c, &contConfig, &hostConfig, nil, nil, adapterName)

	if err != nil {
		panic(err)
	}
	log.Println("[DMS Adapter] Created NuNet Adapter container")
	log.Println("[DMS Adapter] Starting NuNet Adapter container")
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

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
}
