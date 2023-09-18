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
	RewardType        string `json:"reward_type,omitempty"`
	SignatureDatum    string `json:"signature_datum,omitempty"`
	MessageHashDatum  string `json:"message_hash_datum,omitempty"`
	Datum             string `json:"datum,omitempty"`
	SignatureAction   string `json:"signature_action,omitempty"`
	MessageHashAction string `json:"message_hash_action,omitempty"`
	Action            string `json:"action,omitempty"`
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
	if err := db.DB.Where("tx_hash = ?", body.TxHash).Find(&service).Error; err != nil {
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
	oracleResp, err := oracle.WithdrawTokenRequest(&oracle.RewardRequest{
		JobStatus:            service.JobStatus,
		JobDuration:          service.JobDuration,
		EstimatedJobDuration: service.EstimatedJobDuration,
		LogPath:              service.LogURL,
		EstimatedPrice:       service.EstimatedNTX,
		MetadataHash:         service.MetadataHash,
		WithdrawHash:         service.WithdrawHash,
		RefundHash:           service.RefundHash,
		DistributeHash:       service.DistributeHash,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "connetction to oracle failed"})
		return
	}

	resp := rewardRespToCPD{
		RewardType:        oracleResp.RewardType,
		SignatureDatum:    oracleResp.SignatureDatum,
		MessageHashDatum:  oracleResp.MessageHashDatum,
		Datum:             oracleResp.Datum,
		SignatureAction:   oracleResp.SignatureAction,
		MessageHashAction: oracleResp.MessageHashAction,
		Action:            oracleResp.Action,
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

	if body.TransactionType == "withdraw" && body.TransactionStatus == "success" {
		zlog.Sugar().Infof("withdraw transaction successful - deleting from services")
		// Delete the entry
		if err := db.DB.Where("tx_hash = ?", body.TxHash).Delete(&models.Services{}).Error; err != nil {
			zlog.Sugar().Errorln(err)
		}
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("transaction status %s acknowledged", body.TransactionStatus)})
}
