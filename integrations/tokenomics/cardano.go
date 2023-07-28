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

type rewardRespToCPD struct {
	Signature     string `json:"signature,omitempty"`
	OracleMessage string `json:"oracle_message,omitempty"`
	RewardType    string `json:"reward_type,omitempty"`
}

// HandleRequestReward  godoc
//
//	@Summary		Get NTX tokens for work done.
//	@Description	HandleRequestReward takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.
//	@Tags			run
//	@Param			body	body		ClaimCardanoTokenBody	true	"Claim Reward Body"
//	@Success		200		{object}	rewardRespToCPD
//	@Router			/run/request-reward [post]
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
	oracleResp, err := oracle.WithdrawTokenRequest(service)
	if err != nil {
		c.JSON(500, gin.H{"error": "connetction to oracle failed"})
		return
	}

	resp := rewardRespToCPD{
		Signature:     oracleResp.Signature,
		OracleMessage: oracleResp.OracleMessage,
		RewardType:    oracleResp.RewardType,
	}

	c.JSON(200, resp)
}
