package tokenomics

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cosmos/btcutil/bech32"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

// CardanoNodeEndpoint type for Nunet's cardano rest api endpoints
type CardanoNodeEndpoint string

const (
	// CardanoMainnet - mainnet Nunet's cardano rest api endpoint
	CardanoMainnet CardanoNodeEndpoint = "api.cardano.nunet.io"

	// CardanoPreProd - testnet preprod Nunet's cardano rest api endpoint
	CardanoPreProd CardanoNodeEndpoint = "preprod.cardano.nunet.io"

	PreprodContractAddr = "addr_test1wp2nthmgs7n6cwfs4srjs8vtsayhss08he0vwyl2e0v836s5n6jgy"
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

type Amount struct {
	Unit     string `json:"unit"`
	Quantity string `json:"quantity"`
}

type Input struct {
	OutputIndex int    `json:"outputIndex"`
	TxHash      string `json:"txHash"`
}

type Output struct {
	Address string   `json:"address"`
	Amount  []Amount `json:"amount"`
}

type UTXO struct {
	Input  Input  `json:"input"`
	Output Output `json:"output"`
}

// Bytes struct
type Bytes struct {
	Bytes string `json:"bytes"`
}

// Int struct
type Int struct {
	Int int `json:"int"`
}

// Fields struct
type Fields struct {
	Bytes string `json:"bytes,omitempty"`
	Int   int    `json:"int,omitempty"`
}

type InlineDatum struct {
	Constructor int      `json:"constructor"`
	Fields      []Fields `json:"fields"`
}

type UTXOResponse struct {
	UTXOs  []UTXO        `json:"utxo"`
	Datums []InlineDatum `json:"datum"`
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
		c.JSON(404, gin.H{"error": "unknown tx hash"})
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

	utxoHashes, err := GetUTXOsOfSmartContract(body.Address, CardanoPreProd)
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

	err = UpdateTransactionStatus(services, utxoHashes)
	if err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update transaction status"})
	}

	c.JSON(200, gin.H{"message": "Transaction statuses synchronized with blockchain successfully"})
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

// DoesTxExist returns true if the transaction exists in the blockchain
func DoesTxExist(txHash string, endpoint CardanoNodeEndpoint) (bool, error) {
	type Request struct {
		ScriptAddress string `json:"scriptAddress"`
		Env           string `json:"env"`
	}
	reqBody, _ := json.Marshal(Request{ScriptAddress: PreprodContractAddr, Env: "testnet"})

	fullUrl := fmt.Sprintf("http://%s/api/v1/utxo", endpoint)

	resp, err := http.Post(fullUrl, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	res := UTXOResponse{}
	jsonDecoder := json.NewDecoder(resp.Body)
	if err := jsonDecoder.Decode(&res); err != nil && err != io.EOF {
		return false, err
	}

	if len(res.UTXOs) == 0 {
		return false, fmt.Errorf("unable to find receiver")
	}

	for _, utxo := range res.UTXOs {
		if utxo.Input.TxHash == txHash {
			return true, nil
		}
	}

	return false, nil
}

// GetTxReceiver returns the list of receivers of a transaction from the transaction hash
func GetTxReceiver(txHash string, endpoint CardanoNodeEndpoint) (string, error) {
	type Request struct {
		ScriptAddress string `json:"scriptAddress"`
		Env           string `json:"env"`
	}
	reqBody, _ := json.Marshal(Request{ScriptAddress: PreprodContractAddr, Env: "testnet"})

	fullUrl := fmt.Sprintf("http://%s/api/v1/utxo", endpoint)
	resp, err := http.Post(fullUrl, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	res := UTXOResponse{}
	jsonDecoder := json.NewDecoder(resp.Body)
	if err := jsonDecoder.Decode(&res); err != nil && err != io.EOF {
		return "", err
	}

	if len(res.UTXOs) == 0 {
		return "", fmt.Errorf("utxo count is zero, can't find receiver")
	}

	inlineDatum := findDatumForTx(res, txHash)

	if len(inlineDatum.Fields) == 0 {
		return "", fmt.Errorf("inline datum fields count is zero, can't find receiver")
	}

	receiver := inlineDatum.Fields[1].Bytes

	return receiver, nil
}

func findDatumForTx(response UTXOResponse, txHash string) InlineDatum {
	for i, utxo := range response.UTXOs {
		if utxo.Input.TxHash == txHash {
			return response.Datums[i]
		}
	}
	return InlineDatum{}
}

// GetTxConfirmations returns the number of confirmations of a transaction from the transaction hash
//
// Deprecated: We switched from koios to our own cardano node. Use DoesTxExist instead.
func GetTxConfirmations(txHash string, endpoint CardanoNodeEndpoint) (int, error) {
	type Request struct {
		TxHashes []string `json:"_tx_hashes"`
	}
	reqBody, _ := json.Marshal(Request{TxHashes: []string{txHash}})

	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/utxo", endpoint),
		"application/json",
		bytes.NewBuffer(reqBody))

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var res []struct {
		TxHash        string `json:"tx_hash"`
		Confirmations int    `json:"num_confirmations"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return 0, err
	}

	return res[len(res)-1].Confirmations, nil
}

// GetUTXOsOfSmartContract fetch all utxos of smart contract and return list of tx_hash
func GetUTXOsOfSmartContract(address string, endpoint CardanoNodeEndpoint) ([]string, error) {
	type Request struct {
		Address  []string `json:"_addresses"`
		Extended bool     `json:"_extended"`
	}
	reqBody, _ := json.Marshal(Request{Address: []string{address}, Extended: true})

	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/utxo", endpoint),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	res := UTXOResponse{}
	jsonDecoder := json.NewDecoder(resp.Body)
	if err := jsonDecoder.Decode(&res); err != nil && err != io.EOF {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	var utxoHashes []string
	for _, utxo := range res.UTXOs {
		utxoHashes = append(utxoHashes, utxo.Input.TxHash)
	}

	return utxoHashes, nil
}

// WaitForTxConfirmation waits for a transaction to be confirmed
func WaitForTxConfirmation(confirmations int, timeout time.Duration,
	txHash string, endpoint CardanoNodeEndpoint) error {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	confirmationsMade := 0

	for {
		select {
		case <-ticker.C:
			zlog.Sugar().Debugf("inside WaitForTxConfirmation ticker with %d confirmations", confirmations)
			exists, err := DoesTxExist(txHash, endpoint)
			if err != nil {
				continue
			}
			if exists {
				confirmationsMade++
			}
			if confirmationsMade >= confirmations {
				return nil
			}
		case <-time.After(timeout):
			return errors.New("timeout")
		}
	}
}

type UTXOs struct {
	TxHash  string `json:"tx_hash"`
	IsSpent bool   `json:"is_spent"`
}

// isValidCardano checks if the cardano address is valid
func isValidCardano(addr string, valid *bool) {
	defer func() {
		if r := recover(); r != nil {
			*valid = false
		}
	}()
	if _, err := address.NewAddress(addr); err == nil {
		*valid = true
	}
}

// ValidateAddress checks if the wallet address is a valid ethereum/cardano address
func ValidateAddress(addr string) error {
	if common.IsHexAddress(addr) {
		return errors.New("ethereum wallet address not allowed")
	}

	var validCardano = false
	isValidCardano(addr, &validCardano)
	if validCardano {
		return nil
	}

	return errors.New("invalid cardano wallet address")
}

func GetAddressPaymentCredential(addr string) (string, error) {
	_, data, err := bech32.Decode(addr, 1023)
	if err != nil {
		return "", fmt.Errorf("decoding bech32 failed: %w", err)
	}

	converted, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", fmt.Errorf("decoding bech32 failed: %w", err)
	}

	return hex.EncodeToString(converted)[2:58], nil
}

// SliceContains checks if a string exists in a slice
func SliceContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// UpdateTransactionStatus updates the status of claimed transactions in local DB
func UpdateTransactionStatus(services []models.Services, utxoHashes []string) error {
	for _, service := range services {
		if !SliceContains(utxoHashes, service.TxHash) {
			if service.TransactionType == "withdraw" {
				service.TransactionType = transactionWithdrawnStatus
			} else if service.TransactionType == "refund" {
				service.TransactionType = transactionRefundedStatus
			} else if service.TransactionType == "distribute-50" || service.TransactionType == "distribute-75" {
				service.TransactionType = transactionDistributedStatus
			}

			if err := db.DB.Save(&service).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
