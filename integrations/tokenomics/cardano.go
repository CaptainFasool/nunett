package tokenomics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

type TxHashResp struct {
	TxHash          string `json:"tx_hash"`
	TransactionType string `json:"transaction_type"`
	DateTime        string `json:"date_time"`
}
type ClaimCardanoTokenBody struct {
	ComputeProviderAddress string `json:"compute_provider_address"`
	TxHash                 string `json:"tx_hash"`
}
type rewardRespToCPD struct {
	ServiceProviderAddr string `json:"service_provider_addr"`
	ComputeProviderAddr string `json:"compute_provider_addr"`
	RewardType          string `json:"reward_type,omitempty"`
	SignatureDatum      string `json:"signature_datum,omitempty"`
	MessageHashDatum    string `json:"message_hash_datum,omitempty"`
	Datum               string `json:"datum,omitempty"`
	SignatureAction     string `json:"signature_action,omitempty"`
	MessageHashAction   string `json:"message_hash_action,omitempty"`
	Action              string `json:"action,omitempty"`
}
type updateTxStatusBody struct {
	Address string `json:"address,omitempty"`
}

// GetJobTxHashes  godoc
//
//	@Summary		Get list of TxHashes for jobs done.
//	@Description	Get list of TxHashes along with the date and time of jobs done.
//	@Tags			run
//	@Success		200		{object}	TxHashResp
//	@Router			/transactions [get]
func GetJobTxHashes(c *gin.Context) {
	var err error

	sizeDoneStr := c.Query("size_done")
	cleanTx := c.Query("clean_tx")
	sizeDone, err := strconv.Atoi(sizeDoneStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size_done parameter"})
		return
	}

	if cleanTx != "done" && cleanTx != "refund" && cleanTx != "withdraw" && cleanTx != "" {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "wrong clean_tx parameter"})
		return
	}

	if len(cleanTx) > 0 {
		if err = db.DB.Where("transaction_type = ?", cleanTx).Delete(&models.Services{}).Error; err != nil {
			zlog.Sugar().Errorf("%+v", err)
		}
	}

	var txHashesResp []TxHashResp
	var services []models.Services
	if sizeDone == 0 {
		err = db.DB.
			Where("tx_hash IS NOT NULL").
			Where("log_url LIKE ?", "%log.nunet.io%").
			Where("transaction_type is NOT NULL").
			Find(&services).Error
		if err != nil {
			zlog.Sugar().Errorf("%+v", err)
			c.JSON(404, gin.H{"error": "no job deployed to request reward for"})
			return
		}

	} else {
		services, err = getLimitedTransactions(sizeDone)
		if err != nil {
			zlog.Sugar().Errorf("%+v", err)
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
	}

	for _, service := range services {
		txHashesResp = append(txHashesResp, TxHashResp{
			TxHash:          service.TxHash,
			TransactionType: service.TransactionType,
			DateTime:        service.CreatedAt.String(),
		})
	}

	if len(txHashesResp) == 0 {
		c.JSON(404, gin.H{"error": "no job deployed to request reward for"})
		return
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
	var payload ClaimCardanoTokenBody

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		c.Abort()
		return
	}

	// At some point, management dashboard should send container ID to identify
	// against which container we are requesting reward
	service := models.Services{TxHash: payload.TxHash}
	// SELECTs the first record; first record which is not marked as delete
	if err := db.DB.Where("tx_hash = ?", payload.TxHash).Find(&service).Error; err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction hash not found or invalid"})
		return
	}
	zlog.Sugar().Infof("service found from txHash: %+v", service)

	if service.JobStatus == "running" {
		c.JSON(503, gin.H{"error": "the job is still running"})
		return
	}

	resp := rewardRespToCPD{
		ServiceProviderAddr: service.ServiceProviderAddr,
		ComputeProviderAddr: service.ComputeProviderAddr,
		RewardType:          service.TransactionType,
		SignatureDatum:      service.SignatureDatum,
		MessageHashDatum:    service.MessageHashDatum,
		Datum:               service.Datum,
		SignatureAction:     service.SignatureAction,
		MessageHashAction:   service.MessageHashAction,
		Action:              service.Action,
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

	if body.TransactionStatus == "success" {
		zlog.Sugar().Infof("withdraw transaction successful - updating DB")
		// Partial deletion of entry
		var service models.Services
		if err := db.DB.Where("tx_hash = ?", body.TxHash).Find(&service).Error; err != nil {
			zlog.Sugar().Errorln(err)
		}
		service.TransactionType = "done"
		db.DB.Save(&service)
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("transaction status %s acknowledged", body.TransactionStatus)})
}

// HandleUpdateStatus  godoc
//
//	@Summary		Updates blockchain transaction status of DB.
//	@Description	HandleUpdateStatus is used by webapps to update status of saved transactions with fetching info from blockchain using koios REST API.
//	@Tags			tx
//	@Param			body	body		updateTxStatusBody	true	"Transaction Status Update Body"
//	@Success		200		{string}	string
//	@Router			/transactions/update-status [post]
func HandleUpdateStatus(c *gin.Context) {
	body := updateTxStatusBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot read payload body"})
		return
	}

	utxoHashes, err := utils.GetUTXOsOfSmartContract(body.Address, utils.KoiosPreProd)
	if err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch UTXOs from Blockchain"})
		return
	}

	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	var services []models.Services
	err = db.DB.
		Where("tx_hash IS NOT NULL").
		Where("log_url LIKE ?", "%log.nunet.io%").
		Where("transaction_type IS NOT NULL").
		Where("deleted_at IS NULL").
		Where("created_at <= ?", fiveMinutesAgo).
		Not("transaction_type = ?", "done").
		Not("transaction_type = ?", "").
		Find(&services).Error
	if err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "no job deployed to request reward for"})
		return
	}

	err = utils.UpdateTransactionStatus(services, utxoHashes)
	if err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update transaction status"})
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Transaction statuses synchronized with blockchain successfully")})
}

func getLimitedTransactions(sizeDone int) ([]models.Services, error) {
	var doneServices []models.Services
	var services []models.Services
	err := db.DB.
		Where("tx_hash IS NOT NULL").
		Where("log_url LIKE ?", "%log.nunet.io%").
		Where("transaction_type = ?", "done").
		Order("created_at DESC").
		Limit(sizeDone).
		Find(&doneServices).Error
	if err != nil {
		return []models.Services{}, err
	}

	err = db.DB.
		Where("tx_hash IS NOT NULL").
		Where("log_url LIKE ?", "%log.nunet.io%").
		Where("transaction_type IS NOT NULL").
		Not("transaction_type = ?", "done").
		Not("transaction_type = ?", "").
		Find(&services).Error
	if err != nil {
		return []models.Services{}, err
	}

	services = append(services, doneServices...)
	return services, nil
}
