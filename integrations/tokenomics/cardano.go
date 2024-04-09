package tokenomics

import (
	"fmt"
	"time"

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
	TxHash                 string `json:"tx_hash" binding:"required"`
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

type UpdateTxStatusBody struct {
	Address string `json:"address,omitempty" binding:"required"`
}

func GetJobTxHashes(size int, clean string) ([]TxHashResp, error) {
	if clean != "done" && clean != "refund" && clean != "withdraw" && clean != "" {
		return nil, fmt.Errorf("invalid clean_tx parameter")
	}

	err := db.DB.Where("transaction_type = ?", clean).Delete(&models.Services{}).Error
	if err != nil {
		zlog.Sugar().Errorf("%w", err)
	}

	var resp []TxHashResp
	var services []models.Services
	if size == 0 {
		err = db.DB.
			Where("tx_hash IS NOT NULL").
			Where("log_url LIKE ?", "%log.nunet.io%").
			Where("transaction_type is NOT NULL").
			Find(&services).Error
		if err != nil {
			zlog.Sugar().Errorf("%w", err)
			return nil, fmt.Errorf("no job deployed to request reward for: %w", err)
		}
	} else {
		services, err = getLimitedTransactions(size)
		if err != nil {
			zlog.Sugar().Errorf("%w", err)
			return nil, fmt.Errorf("could not get limited transactions: %w", err)
		}
	}
	for _, service := range services {
		resp = append(resp, TxHashResp{
			TxHash:          service.TxHash,
			TransactionType: service.TransactionType,
			DateTime:        service.CreatedAt.String(),
		})
	}
	return resp, nil
}

func RequestReward(claim ClaimCardanoTokenBody) (*rewardRespToCPD, error) {
	// At some point, management dashboard should send container ID to identify
	// against which container we are requesting reward
	service := models.Services{
		TxHash: claim.TxHash,
	}

	// SELECTs the first record; first record which is not marked as delete
	err := db.DB.Where("tx_hash = ?", claim.TxHash).Find(&service).Error
	if err != nil {
		zlog.Sugar().Errorln(err)
		return nil, fmt.Errorf("unknown tx hash: %w", err)
	}

	zlog.Sugar().Infof("service found from txHash: %+v", service)
	if service.JobStatus == "running" {
		return nil, fmt.Errorf("job is still running")
		// c.JSON(503, gin.H{"error": "the job is still running"})
	}

	reward := rewardRespToCPD{
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
	return &reward, nil
}

func SendStatus(status models.BlockchainTxStatus) string {
	if status.TransactionStatus == "success" {
		zlog.Sugar().Infof("withdraw transaction successful - updating DB")
		// Partial deletion of entry
		var service models.Services
		err := db.DB.Where("tx_hash = ?", status.TxHash).Find(&service).Error
		if err != nil {
			zlog.Sugar().Errorln(err)
		}
		service.TransactionType = "done"
		db.DB.Save(&service)
	}
	return status.TransactionStatus
}

func UpdateStatus(body UpdateTxStatusBody) error {
	utxoHashes, err := utils.GetUTXOsOfSmartContract(body.Address, utils.KoiosPreProd)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to fetch UTXOs from Blockchain: %w", err)
	}

	fiveMinAgo := time.Now().Add(-5 * time.Minute)
	var services []models.Services
	err = db.DB.
		Where("tx_hash IS NOT NULL").
		Where("log_url LIKE ?", "%log.nunet.io%").
		Where("transaction_type IS NOT NULL").
		Where("deleted_at IS NULL").
		Where("created_at <= ?", fiveMinAgo).
		Not("transaction_type = ?", "done").
		Not("transaction_type = ?", "").
		Find(&services).Error
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("no job deployed to request reward for: %w", err)
	}

	err = utils.UpdateTransactionStatus(services, utxoHashes)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to update transaction status")
	}
	return nil
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
