package utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cosmos/btcutil/bech32"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fivebinaries/go-cardano-serialization/address"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

// KoiosEndpoint type for Koios rest api endpoints
type KoiosEndpoint string

const (
	// KoiosMainnet - mainnet Koios rest api endpoint
	KoiosMainnet KoiosEndpoint = "api.koios.rest"

	// KoiosPreProd - testnet preprod Koios rest api endpoint
	KoiosPreProd KoiosEndpoint = "preprod.koios.rest"
)

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

// GetTxReceiver returns the list of receivers of a transaction from the transaction hash
func GetTxReceiver(txHash string, endpoint KoiosEndpoint) (string, error) {
	type Request struct {
		TxHashes []string `json:"_tx_hashes"`
	}
	reqBody, _ := json.Marshal(Request{TxHashes: []string{txHash}})

	resp, err := http.Post(
		fmt.Sprintf("https://%s/api/v1/tx_info", endpoint),
		"application/json",
		bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	res := []struct {
		Outputs []struct {
			InlineDatum struct {
				Value struct {
					Fields []struct {
						Bytes string `json:"bytes"`
					} `json:"fields"`
				} `json:"value"`
			} `json:"inline_datum"`
		} `json:"outputs"`
	}{}
	jsonDecoder := json.NewDecoder(resp.Body)
	if err := jsonDecoder.Decode(&res); err != nil && err != io.EOF {
		return "", err
	}

	if len(res) == 0 || len(res[0].Outputs) == 0 || len(res[0].Outputs[1].InlineDatum.Value.Fields) == 0 {
		return "", fmt.Errorf("unable to find receiver")
	}

	receiver := res[0].Outputs[1].InlineDatum.Value.Fields[1].Bytes

	return receiver, nil
}

// GetTxConfirmations returns the number of confirmations of a transaction from the transaction hash
func GetTxConfirmations(txHash string, endpoint KoiosEndpoint) (int, error) {
	type Request struct {
		TxHashes []string `json:"_tx_hashes"`
	}
	reqBody, _ := json.Marshal(Request{TxHashes: []string{txHash}})

	resp, err := http.Post(
		fmt.Sprintf("https://%s/api/v1/tx_status", endpoint),
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

// WaitForTxConfirmation waits for a transaction to be confirmed
func WaitForTxConfirmation(confirmations int, timeout time.Duration, txHash string, endpoint KoiosEndpoint) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conf, err := GetTxConfirmations(txHash, endpoint)
			if err != nil {
				return err
			}
			if conf >= confirmations {
				return nil
			}
		case <-time.After(timeout):
			return errors.New("timeout")
		}
	}
}

// GetUTXOsOfSmartContract fetch all utxos of smart contract and return list of tx_hash
func GetUTXOsOfSmartContract(address string, endpoint KoiosEndpoint) ([]string, error) {
	type Request struct {
		Address  []string `json:"_addresses"`
		Extended bool     `json:"_extended"`
	}
	reqBody, _ := json.Marshal(Request{Address: []string{address}, Extended: true})

	resp, err := http.Post(
		fmt.Sprintf("https://%s/api/v1/address_utxos", endpoint),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	var utxos []UTXOs
	jsonDecoder := json.NewDecoder(resp.Body)
	if err := jsonDecoder.Decode(&utxos); err != nil && err != io.EOF {
		return nil, err
	}

	var utxoHashes []string
	for _, utxo := range utxos {
		utxoHashes = append(utxoHashes, utxo.TxHash)
	}

	return utxoHashes, nil
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
