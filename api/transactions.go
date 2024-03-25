package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/integrations/tokenomics"
	"gitlab.com/nunet/device-management-service/models"
)

// GetJobTxHashesHandler  godoc
//
//	@Summary		Get list of TxHashes for jobs done.
//	@Description	Get list of TxHashes along with the date and time of jobs done.
//	@Tags			run
//	@Success		200		{object}	[]tokenomics.TxHashResp
//	@Router			/transactions [get]
func GetJobTxHashesHandler(c *gin.Context) {
	sizeStr := c.Query("size_done")
	clean := c.Query("clean_tx")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid size_done parameter"})
		return
	}
	hashes, err := tokenomics.GetJobTxHashes(size, clean)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, hashes)
}

// RequestRewardHandler  godoc
//
//	@Summary		Get NTX tokens for work done.
//	@Description	HandleRequestReward takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.
//	@Tags			run
//	@Param			body	body		tokenomics.ClaimCardanoTokenBody	true	"Claim Reward Body"
//	@Success		200		{object}	tokenomics.rewardRespToCPD
//	@Router			/transactions/request-reward [post]
func RequestRewardHandler(c *gin.Context) {
	var payload tokenomics.ClaimCardanoTokenBody
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}
	resp, err := tokenomics.RequestReward(payload)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, resp)
}

// SendTxStatusHandler  godoc
//
//	@Summary		Sends blockchain status of contract creation.
//	@Description	HandleSendStatus is used by webapps to send status of blockchain activities. Such token withdrawl.
//	@Tags			run
//	@Param			body	body		models.BlockchainTxStatus	true	"Blockchain Transaction Status Body"
//	@Success		200		{string}	string
//	@Router			/transactions/send-status [post]
func SendTxStatusHandler(c *gin.Context) {
	body := models.BlockchainTxStatus{}
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "cannot read payload body"})
		return
	}
	status := tokenomics.SendStatus(body)
	c.JSON(200, gin.H{"message": fmt.Sprintf("transaction status %s acknowledged", status)})
}

// UpdateTxStatusHandler  godoc
//
//	@Summary		Updates blockchain transaction status of DB.
//	@Description	HandleUpdateStatus is used by webapps to update status of saved transactions with fetching info from blockchain using koios REST API.
//	@Tags			tx
//	@Param			body	body		tokenomics.UpdateTxStatusBody	true	"Transaction Status Update Body"
//	@Success		200		{string}	string
//	@Router			/transactions/update-status [post]
func UpdateTxStatusHandler(c *gin.Context) {
	body := tokenomics.UpdateTxStatusBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid payload data"})
		return
	}
	err = tokenomics.UpdateStatus(body)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "transaction statuses synchronized with blockchain successfully"})
}
