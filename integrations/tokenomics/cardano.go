package tokenomics

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/models"
)

type TxHashResp struct {
	TxHash   string `json:"tx_hash"`
	DateTime string `json:"date_time"`
}
type ClaimCardanoTokenBody struct {
	ComputeProviderAddress string `json:"compute_provider_address"`
	TxHash                 string `json:"tx_hash"`
}
type rewardRespToCPD struct {
	Signature     string `json:"signature,omitempty"`
	OracleMessage string `json:"oracle_message,omitempty"`
	RewardType    string `json:"reward_type,omitempty"`
}

// GetJobTxHashes  godoc
//
//	@Summary		Get list of TxHashes for jobs done.
//	@Description	Get list of TxHashes along with the date and time of jobs done.
//	@Tags			run
//	@Success		200		{object}	TxHashResp
//	@Router			/transactions [get]
func GetJobTxHashes(c *gin.Context) {
	var services []models.Services
	if err := db.DB.Where("tx_hash IS NOT NULL").Where("log_url LIKE ?", "%log.nunet.io%").Find(&services).Error; err != nil {
		zlog.Sugar().Errorf("%+v", err)
		c.JSON(404, gin.H{"error": "no job deployed to request reward for"})
		return
	}

	if len(services) == 0 {
		c.JSON(404, gin.H{"error": "no job deployed to request reward for"})
		return
	}

	var txHashesResp []TxHashResp
	for _, service := range services {
		txHashesResp = append(txHashesResp, TxHashResp{TxHash: service.TxHash, DateTime: service.CreatedAt.String()})
	}

	c.JSON(200, txHashesResp)
}

// HandleRequestReward  godoc
//
//	@Summary		Get NTX tokens for work done.
//	@Description	HandleRequestReward takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.
//	@Tags			run
//	@Param			body	body		ClaimCardanoTokenBody	true	"Claim Reward Body"
//	@Success		200		{object}	rewardRespToCPD
//	@Router			/transactions/request-reward [post]
func HandleRequestReward(c *gin.Context) {
	body := ClaimCardanoTokenBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	// At some point, management dashboard should send container ID to identify
	// against which container we are requesting reward
	service := models.Services{TxHash: body.TxHash}
	// SELECTs the first record; first record which is not marked as delete
	if err := db.DB.Find(&service).Error; err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(404, gin.H{"error": "unknown tx hash"})
		return
	}
	zlog.Sugar().Infof("service found from txHash: %+v", service)

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

// HandleSendStatus  godoc
//
//	@Summary		Sends blockchain status of contract creation.
//	@Description	HandleSendStatus is used by webapps to send status of blockchain activities. Such token withdrawl.
//	@Tags			run
//	@Param			body	body		models.BlockchainTxStatus	true	"Blockchain Transaction Status Body"
//	@Success		200		{string}	string
//	@Router			/transactions/send-status [post]
func HandleSendStatus(c *gin.Context) {
	body := models.BlockchainTxStatus{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot read payload body"})
		return
	}

	var requestTracker models.RequestTracker
	res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	}
	if body.TransactionType == "withdraw" && body.TransactionStatus == "success" {
		zlog.Sugar().Infof("withdraw transaction successful - deleting from services")
		// Delete the entry
		if err := db.DB.Where("tx_hash = ?", body.TxHash).Delete(&models.Services{}).Error; err != nil {
			zlog.Sugar().Errorln(err)
		}
	}

	serviceStatus := body.TransactionType + " with " + body.TransactionStatus

	requestTracker.Status = serviceStatus
	db.DB.Save(&requestTracker)

	c.JSON(200, gin.H{"message": fmt.Sprintf("transaction status %s acknowledged", body.TransactionStatus)})
}
