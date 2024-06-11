package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

// GetJobTxHashesHandler  godoc
//
//	@Summary		Get list of TxHashes for jobs done.
//	@Description	Get list of TxHashes along with the date and time of jobs done.
//	@Tags			transactions
//	@Success		200			{object}	[]tokenomics.TxHashResp
//	@Param			size_done	query		int		false	"Number of transactions to fetch"
//	@Param			clean_tx	query		string	false	"Clean transactions of type"
//	@Failure		500			{object}	string	"no job deployed to request reward for"
//	@Failure		500			{object}	string	"could not get limited transactions"
//	@Failure		400			{object}	string	"invalid size_done parameter"
//	@Failure		400			{object}	string	"wrong clean_tx parameter"
//	@Failure		202			{object}	string	"job is still running"
//	@Router			/transactions [get]
func GetJobTxHashesHandler(c *gin.Context) {
	sizeStr := c.Query("size_done")
	clean := c.Query("clean_tx")
	if clean != "done" && clean != "refund" && clean != "withdraw" && clean != "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "wrong clean_tx parameter"})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid size_done parameter"})
		return
	}
	hashes, err := utils.GetJobTxHashes(size, clean)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, hashes)
}

// RequestRewardHandler  godoc
//
//	@Summary		Get NTX tokens for work done.
//	@Description	RequestRewardHandler takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.
//	@Tags			transactions
//	@Param			body	body		tokenomics.ClaimCardanoTokenBody	true	"Claim Reward Body"
//	@Success		200		{object}	tokenomics.RewardRespToCPD
//	@Failure		400		{object}	object	"invalid request body"
//	@Failure		500		{object}	object	"unknown tx hash"
//	@Failure		500		{object}	object	"the job is still running"
//	@Router			/transactions/request-reward [post]
func RequestRewardHandler(c *gin.Context) {
	var payload utils.ClaimCardanoTokenBody
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := utils.RequestReward(payload)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, resp)
}

// SendTxStatusHandler  godoc
//
//	@Summary		Sends blockchain status of contract creation.
//	@Description	SendTxStatusHandler is used by webapps to send status of blockchain activities. Such token withdrawl.
//	@Tags			transactions
//	@Param			body	body		models.BlockchainTxStatus	true	"Blockchain Transaction Status Body"
//	@Success		200		{string}	string						"transaction status acknowledged"
//	@Failure		400		{object}	string						"cannot read payload body"
//	@Router			/transactions/send-status [post]
func SendTxStatusHandler(c *gin.Context) {
	body := models.BlockchainTxStatus{}
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "cannot read payload body"})
		return
	}
	status := utils.SendStatus(body)
	c.JSON(200, gin.H{"message": fmt.Sprintf("transaction status %s acknowledged", status)})
}

// UpdateTxStatusHandler  godoc
//
//	@Summary		Updates blockchain transaction status of DB.
//	@Description	UpdateTxStatusHandler is used by webapps to update status of saved transactions with fetching info from blockchain using koios REST API.
//	@Tags			transactions
//	@Param			body	body		tokenomics.UpdateTxStatusBody	true	"Transaction Status Update Body"
//	@Success		200		{string}	string							"Transaction statuses synchronized with blockchain successfully"
//	@Failure		400		{object}	string							"cannot read payload body"
//	@Failure		500		{object}	string							"no job deployed to request reward for"
//	@Failure		500		{object}	string							"failed to fetch UTXOs from Blockchain"
//	@Failure		500		{object}	string							"failed to update transaction status"
//	@Router			/transactions/update-status [post]
func UpdateTxStatusHandler(c *gin.Context) {
	body := utils.UpdateTxStatusBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid payload data"})
		return
	}
	err = utils.UpdateStatus(body)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "transaction statuses synchronized with blockchain successfully"})
}
