package tokenomics

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/statsdb"
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
	// SELECTs the first record; first record which is not marked as delete
	if err := db.DB.First(&service).Error; err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(404, gin.H{"error": "no job deployed to request reward for"})
		return
	}

	if service.JobStatus == "running" {
		c.JSON(503, gin.H{"error": "the job is still running"})
		return
	}

	// Send the service data to oracle for examination
	resp, err := oracle.WithdrawTokenRequest(service)
	if err != nil {
		c.JSON(500, gin.H{"error": "connetction to oracle failed"})
		return
	}

	var requestTracker models.RequestTracker
	res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	}

	// sending ntx_payment info to stats database via grpc Call
	NtxPaymentParams := models.NtxPayment{
		CallID:      requestTracker.CallID,
		ServiceID:   requestTracker.ServiceType,
		AmountOfNtx: requestTracker.MaxTokens,
		PeerID:      requestTracker.NodeID,
		Timestamp:   float32(statsdb.GetTimestamp()),
	}
	statsdb.NtxPayment(NtxPaymentParams)

	c.JSON(200, resp)
}
