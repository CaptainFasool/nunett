package tokenomics

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/models"
)

type ClaimCardanoTokenBody struct {
	ComputeProviderAddress string `json:"compute_provider_address"`
}

// HandleRequestReward  godoc
// @Summary      Get NTX tokens for work done.
// @Description  HandleRequestReward takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.
// @Router       /run/request-reward [post]
func HandleRequestReward(c *gin.Context) {
	rand.Seed(time.Now().Unix())

	body := ClaimCardanoTokenBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// At some point, management dashboard should send container ID to identify
	// against which container we are requesting reward
	var service models.Services
	db.DB.First(&service, "id = ?", 1) // Currently we are assuming only 1 row in services table

	if service.JobStatus == "running" {
		c.JSON(102, gin.H{"message": "the job is still running"})
	}

	// Send the service data to oracle for examination
	resp, err := oracle.WithdrawTokenRequest(service)

	if err != nil {
		c.JSON(500, gin.H{"message": "connetction to oracle failed"})
		return
	}

	c.JSON(200, resp)
}