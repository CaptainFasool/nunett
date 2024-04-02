package api

import (
	"fmt"
	"net/http"

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
	var queries struct {
		SizeDone int    `json:"size_done" binding:"number"`
		CleanTx  string `json:"clean_tx"`
	}
	err := c.ShouldBindQuery(&queries)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}
	hashes, err := tokenomics.GetJobTxHashes(queries.SizeDone, queries.CleanTx)
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
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	var claim tokenomics.ClaimCardanoTokenBody
	err := c.ShouldBindJSON(&claim)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}
	resp, err := tokenomics.RequestReward(claim)
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
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	var body models.BlockchainTxStatus
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
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
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	var status tokenomics.UpdateTxStatusBody
	err := c.ShouldBindJSON(&status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}
	err = tokenomics.UpdateStatus(status)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "transaction statuses synchronized with blockchain successfully"})
}
