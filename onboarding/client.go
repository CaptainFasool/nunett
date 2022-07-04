package onboarding

import (
	"net/http"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/models"
)

func CreateClientConfig(c *gin.Context, metadata *models.MetadataV2,
	capacityForNunet *models.CapacityForNunet, hostname string) {
	// create client config
	clientData := struct {
		Name           string
		Network        string
		ReservedCPU    int64
		ReservedMemory int64
	}{
		Name:           hostname,
		Network:        capacityForNunet.Channel,
		ReservedCPU:    metadata.Reserved.CPU,
		ReservedMemory: metadata.Reserved.Memory,
	}

	clientFile, err := os.Create("/etc/nunet/clientV2.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	tmpl, err := template.New("client").Parse(models.ClientTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = tmpl.Execute(clientFile, clientData)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
}
