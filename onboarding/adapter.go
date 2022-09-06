package onboarding

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"text/template"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/models"
)

func CreateAdapterConfig(c *gin.Context, metadata *models.MetadataV2, cardanoPassive, hostname string) {
	// create adapter-definition.json
	var adapterPrefix string
	var dockerImageTag string
	var deploymentType string
	var tokenomicsApiName string
	if metadata.Network == "nunet-development" {
		dockerImageTag = "test"
		adapterPrefix = "testing-nunet-adapter"
		deploymentType = "test"
		tokenomicsApiName = "testing-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			dockerImageTag = "arm-test"
		}
	}
	if metadata.Network == "nunet-private-alpha" {
		dockerImageTag = "latest"
		adapterPrefix = "nunet-adapter"
		deploymentType = "prod"
		tokenomicsApiName = "tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			dockerImageTag = "arm-latest"
		}
	}
	adapterData := struct {
		Datacenters       string
		AdapterPrefix     string
		ClientName        string
		DockerTag         string
		DeploymentType    string
		TokenomicsApiName string
		Cardano           string
	}{
		Datacenters:       metadata.Network,
		AdapterPrefix:     adapterPrefix,
		ClientName:        hostname,
		DockerTag:         dockerImageTag,
		DeploymentType:    deploymentType,
		TokenomicsApiName: tokenomicsApiName,
		Cardano:           cardanoPassive,
	}

	adapterFile, err := os.Create("/etc/nunet/adapter-definitionV2.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	tmpl, err := template.New("adapter").Parse(models.AdapterTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = tmpl.Execute(adapterFile, adapterData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
}

func RunNomadJob(c *gin.Context, jobName string) {
	// add a check if nomad is available on the system
	cmd := exec.Command("nomad", "-v")
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "nomad does not exists. https://gitlab.com/nunet/device-management-service/-/wikis/Installation-Guide"})
		return
	}

	byteValue, err := ioutil.ReadFile("/etc/nunet/adapter-definitionV2.json")

	if err != nil {
		panic(err)
	}

	var jsonContent map[string]interface{}
	json.Unmarshal([]byte(byteValue), &jsonContent)

	// initialize http client
	client := &http.Client{}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://nomad.nunet.io:4646/v1/job/"+jobName, bytes.NewBuffer(byteValue))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}

}
