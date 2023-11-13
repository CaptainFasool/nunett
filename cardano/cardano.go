// This package has some basic functionality to interact with the cardano block chain and the Escrow smart contract
// it currently assumed preprod.
//
// You must have a running cardano-node synchronized on preprod and the TesterAddress must have mNTX and tADA for
// the tests to run properly.
//
// You must also have cardano-cli available on your PATH

package cardano

import (
	"fmt"
	"log"
	"strings"
	"bufio"
	"os/exec"
	"encoding/json"
)

// Address of the testing account, corresponds to tester.addr.
const TesterAddress = "addr_test1vzgxkngaw5dayp8xqzpmajrkm7f7fleyzqrjj8l8fp5e8jcc2p2dk"

// Current alpha preprod contract.
const CurrentContract = "addr_test1wp9pc08wneh5nk5cqdjj7h5vr2905k8sfdjqr9c95etults8xaeud"

// Current alpha preprod NTX native asset.
const mNTX = "8cafc9b387c9f6519cacdce48a8448c062670c810d8da4b232e56313.6d4e5458"

// JSONInput is a intermediate unmarshalled format for custom unmarshalling logic for Inputs
// as Inputs do not have enough.
type JSONInput struct {
	Value map[string]interface{} `json:"value"`
}

// Inputs to a transaction.
type Input struct {
	Key string // The key is the '<transaction-hash>#<index>'.
	Value map[string]int64 // Value is a map from '<policy-id>.<hex-asset-name>' or 'lovelace' to token amount.
}

// An Output to be created by a transaction
type Output struct {
	To string  // To Address aka who will own the output
	Value map[string]int64  // Value is a map from '<policy-id>.<hex-asset-name>' or 'lovelace' to token amount.
}

// A Cardano Transaction.
type Transaction struct {
	Inputs []Input // The inputs to the transaction.
	Outputs map[string]Output // The outputs created by this transaction.
	ChangeAddress string // The address that should be given change from this transaction.
}

// Get the Inputs owned by an address.
func GetUTXOs(address string) []Input{
	cmd := exec.Command ("cardano-cli",
		"query",
       	"utxo",
       	"--address",
       	address,
       	"--out-file",
       	"/dev/stdout",
       	"--testnet-magic",
		"1");

	bytes, err := cmd.Output();
	if (err != nil) {
		log.Fatal(err);
	}

	var dev map[string]JSONInput
	err = json.Unmarshal(bytes, &dev)

	var result []Input

	for key, input := range dev {
		var newInput Input
		newInput.Key = key
		newInput.Value = make(map[string]int64)

		for policy, value := range input.Value {
			if (policy == "lovelace") {
				num := value.(float64)
				newInput.Value["lovelace"] = int64(num)
			} else {
				assets := value.(map[string]interface{})
				for assetHex, amount := range assets {
					newInput.Value[fmt.Sprintf("%s.%s", policy, assetHex)] = int64(amount.(float64))
				}
			}
		}

		result = append(result, newInput)
	}

	return result
}

// Pay to the current escrow smart contract an amount in NTX
func PayToScript( ntx int64 ) (hash string){
	outputs := GetUTXOs(TesterAddress)

	transaction := Transaction{
		Inputs: outputs,
		Outputs: make(map[string]Output),
		ChangeAddress: TesterAddress,
	}

	transaction.Outputs[CurrentContract] = Output{
		To: CurrentContract,
		Value: make(map[string]int64),
	}

	transaction.Outputs[CurrentContract].Value[mNTX] = ntx

	// Set the min ada to be held in the output
	transaction.Outputs[CurrentContract].Value["lovelace"] = 2000000;

	BuildTransaction(transaction)
	SignTransaction()
	hash = GetTransactionHash()
	SubmitTransaction()
	return hash
}

func SubmitTransaction ( ) {
	cmd := exec.Command("cardano-cli",
		"transaction",
		"submit",
		"--tx-file",
		"tx.signed",
		"--testnet-magic",
		"1",
	)

	cmd.Run()
}

// Helper to build a valid transaction
func BuildTransaction( tx Transaction ) {
	// The first build is for estimation, we pass in a wide enough lovelace value so the
	// size is approximated correctly.
	BuildTransactionRaw(tx, 200000)
	EnsureProtocolParameters()
	fee := EstimateFee(tx)
	BalanceTransaction(&tx, fee)
	BuildTransactionRaw(tx, fee)
}

// Sign a transaction with the TesterAddress.
func SignTransaction () {
	cmd := exec.Command("cardano-cli",
		"transaction",
		"sign",
		"--tx-body-file",
		"tx.draft",
		"--signing-key-file",
		"tester.sk",
		"--out-file",
		"tx.signed",
	)

	cmd.Run()
}

// Estimate fee of transaction
func EstimateFee( tx Transaction ) (fee int64) {
	cmd := exec.Command("cardano-cli",
		"transaction",
		"calculate-min-fee",
		"--tx-body-file",
		"tx.draft",
		"--tx-in-count",
		fmt.Sprintf("%d", len(tx.Inputs)),
		"--tx-out-count",
		fmt.Sprintf("%d", len(tx.Outputs)),
		"--witness-count",
		"1",
		"--protocol-params-file",
		"protocol.json",
		"--testnet-magic",
		"42",
		)

	bytes, err := cmd.Output();
	if (err != nil) {
		log.Fatal(err);
	}

	_, fee_err := fmt.Sscan(string(bytes), &fee)
	if (err != nil) {
		log.Fatal(fee_err)
	}

	return fee
}

func BalanceTransaction(tx *Transaction, fee int64) {
	var change map[string]int64 = make(map[string]int64)

	// Seed change with inputs
	for _, input := range tx.Inputs {
		for token, amount := range input.Value {
			if (change[token] == -1) {
				change[token] = amount
			} else {
				change[token] += amount
			}
		}
	}

	// Subtract outputs
	for _, output := range tx.Outputs {
		for token, amount := range output.Value {
			change[token] -= amount
		}
	}

	// Deduct fees
	change["lovelace"] -= fee

	// Add entry if it doesn't exist yet
	entry := tx.Outputs[tx.ChangeAddress];
	if entry.Value == nil {
		entry.To = tx.ChangeAddress
		entry.Value = make(map[string]int64);
		tx.Outputs[tx.ChangeAddress] = entry
	}

	// Apply to ChangeAddress output
	for token, amount := range change {
		if (tx.Outputs[tx.ChangeAddress].Value[token] == -1) {
			tx.Outputs[tx.ChangeAddress].Value[token] = amount
		} else {
			tx.Outputs[tx.ChangeAddress].Value[token] += amount
		}
	}
}

// Ensures a valid protocol.json file exists and creates one from the connected chain otherwise.
func EnsureProtocolParameters() {
	cmd := exec.Command("cardano-cli",
		"query",
		"protocol-parameters",
		"--out-file",
		"protocol.json",
		"--testnet-magic",
		"1")

	cmd.Run()
}

// Get the transaction hash of the most recently built transaction.
func GetTransactionHash() (hash string) {
	cmd := exec.Command("cardano-cli", "transaction", "txid", "--tx-file", "tx.signed")

	bytes, err := cmd.Output();
	if (err != nil) {
		log.Fatal(err);
	}

	hash = strings.Trim(string(bytes), "\n")
	return
}

// Builds a transaction file with the given fee.
func BuildTransactionRaw( tx Transaction, fee int64 ) {
	args := make([]string, 0)

	args = append(args,
		"transaction",
		"build-raw",
		"--fee",
		fmt.Sprintf("%d", fee),
		"--out-file",
		"tx.draft",
	)

	for _, input := range tx.Inputs {
		args = append(args, "--tx-in")
		args = append(args, input.Key)
	}

	for _, output := range tx.Outputs {
		args = append(args, "--tx-out")
		args = append(args, fmt.Sprintf(`%s+%s`, output.To, ValueStr(output.Value)))
	}

	cmd := exec.Command ("cardano-cli", args...)

	pipe, _ := cmd.StderrPipe()
    scanner := bufio.NewScanner(pipe)

	go func() {
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()

	cmd.Start()
	cmd.Wait()
}

// Create a cardano-cli compatible multi-asset vaule string from a map of asset to amount
func ValueStr( value map[string]int64 ) (str string) {
	builder := strings.Builder{}

	count := 0
	output_count := len(value)
	for token, amount := range value {
		if (token == "lovelace") {
			builder.WriteString(fmt.Sprintf("%d", amount))
		} else {
			builder.WriteString(fmt.Sprintf("%d %s", amount, token))
		}

		if (count < output_count - 1) {
			builder.WriteString("+")
		}

		count += 1
	}

	return builder.String()
}
