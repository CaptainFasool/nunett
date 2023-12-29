// This package has some basic functionality to interact with the cardano block chain and the Escrow smart contract
// it currently assumed preprod.
//
// You must have a running cardano-node synchronized on preprod and the SPAddress must have mNTX and tADA for
// the tests to run properly.
//
// You must also have cardano-cli available on your PATH

package cardano

import (
	"bytes"
	"fmt"
	"time"
	"log"
	"errors"
	"os"
	"strings"
	"bufio"
	"regexp"
	"os/exec"
	"encoding/json"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"encoding/hex"
)

// NOTE: This corresponds to
// https://gitlab.com/nunet/tokenomics-api/tokenomics-api-cardano-v2/-/blob/develop/src/SimpleEscrow.hs?ref_type=heads#L103
// and must match exactly with the deployed contract
type Redeemer int
const (
	Withdraw Redeemer = 0
	Refund Redeemer = 1
	Distribute Redeemer = 2
)

const (
	testnetMagic = "1"

	SPAccount = "tester"
	CPAccount = "cp"

	// Address of the testing account, corresponds to tester.addr.
	SPAddress = "addr_test1vzgxkngaw5dayp8xqzpmajrkm7f7fleyzqrjj8l8fp5e8jcc2p2dk"
	CPAddress = "addr_test1vz7j9m50apx99phkat3h4sm4t82wvap65ezjpe5xl3kcuxcr6e44h"

	SPCollateral = "1ad86a8ef29adf0aeac0a9c0282a2ee5bc9a3f9a3fc8e8b1c3f086b679aed882#0"
	CPCollateral = "eb455c103d609e7b9d16056486c8203c59a7e7af312f9ecf9b5b6afc899ab71a#0"

	// Current alpha preprod contract.
	CurrentContract = "addr_test1wp2nthmgs7n6cwfs4srjs8vtsayhss08he0vwyl2e0v836s5n6jgy"

	// Current alpha preprod NTX native asset.
	mNTX = "8cafc9b387c9f6519cacdce48a8448c062670c810d8da4b232e56313.6d4e5458"

	// Oracle given metadata hash and withdraw hash, signed with the nunet-team channel oracle
	PreGenMetaDataHash = "dd00ba663a650bcd03f54682a2585da7488a452047f3b515878fcf2379e2ba28cc56bca6f20ff3adefae6072c59bdf288869bc423eb1119f25cd001493e1e505"
	PreGenWithdrawHash = "66756e64696e672d622758205c7865655c786261607e5c786464255c7864325c7831655c7831385c7861615c7839645c7865385c786362756d5c7862615c7861665c783130545c7866315c7839645c7862375c7830365c7864335c78303829425c78646569556f5c78653527"

	DATUM_FORMAT_STRING = `{
      "constructor": 0,
      "fields": [
         {
            "bytes": "%s"
         },
         {
            "bytes": "%s"
         },
         {
            "int": %d
         },
         {
            "bytes": "%s"
         },
         {
            "bytes": "%s"
         },
         {
            "bytes": ""
         },
         {
            "bytes": ""
         }
      ]
   }`

   REDEEMER_FORMAT_STRING = `{
     "constructor": %d,
     "fields": [
      {
   		"constructor": 0,
   		"fields": [
   			{
   				"bytes" : "%s"
   			},
   			{
   				"bytes" : "%s"
   			},
   			{
   				"bytes" : "%s"
   			},
   			{
   				"bytes" : "%s"
   			},
   			{
   				"bytes" : "%s"
   			},
   			{
   				"bytes" : "%s"
   			}
   	 	]
   	  }
     ]
   }`

)

// JSONInput is a intermediate unmarshalled format for custom unmarshalling logic for Inputs
// as Inputs do not have enough.
type JSONInput struct {
	Value map[string]interface{} `json:"value"`
}

// Inputs to a transaction.
type Input struct {
	Key string // The key is the '<transaction-hash>#<index>'.
	Value map[string]int64 // Value is a map from '<policy-id>.<hex-asset-name>' or 'lovelace' to token amount.
	ScriptFile string
	DatumFile string
	RedeemerFile string
}

// An Output to be created by a transaction
type Output struct {
	To string  // To Address aka who will own the output
	Value map[string]int64  // Value is a map from '<policy-id>.<hex-asset-name>' or 'lovelace' to token amount.
	DatumFile string // This is a json file encoding the datum that will be used or nil for no datum
}

// A Cardano Transaction.
type Transaction struct {
	Inputs []Input // The inputs to the transaction.
	Outputs map[string]Output // The outputs created by this transaction.
	ChangeAddress string // The address that should be given change from this transaction.
	Collateral string
}

// Get the Inputs owned by an address.
func GetUTXOs(address string) ([]Input, error){
	cmd := exec.Command ("cardano-cli",
		"query",
       	"utxo",
       	"--address",
       	address,
       	"--out-file",
       	"/dev/stdout",
       	"--testnet-magic",
		testnetMagic);
	output, err := execCommand(cmd)

	var result []Input

	if err != nil {
		return result, err
	}

	var dev map[string]JSONInput
	json.Unmarshal([]byte(output), &dev)

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

	return result, err
}

// Pay to the current escrow smart contract an amount in NTX
func PayToScript( ntx int64, spPubKey string, cpPubKey string ) (string, error){
	outputs, err := GetUTXOs(SPAddress)
	if err != nil {
		return "", err
	}

	transaction := Transaction{
		Inputs: outputs,
		Outputs: make(map[string]Output),
		ChangeAddress: SPAddress,
	}

	const datumFile = "datum.json"
	WriteDatumFile(datumFile, ntx, spPubKey, cpPubKey)

	transaction.Outputs[CurrentContract] = Output{
		To: CurrentContract,
		Value: make(map[string]int64),
		DatumFile: datumFile,
	}

	transaction.Outputs[CurrentContract].Value[mNTX] = ntx

	// Set the min ada to be held in the output
	transaction.Outputs[CurrentContract].Value["lovelace"] = minLovelace

	BuildTransaction(transaction)
	SignTransaction()
	hash, err := GetTransactionHash()
	if err != nil {
		return "", err
	}
	err = SubmitTransaction()
	return hash, err
}

func execCommand(cmd *exec.Cmd) (string, error) {
  var stdout, stderr bytes.Buffer
  cmd.Stdout = &stdout
  cmd.Stderr = &stderr
  err := cmd.Run()
  if err != nil {
		fmt.Println("Command:", cmd.String()) // Print the command
		fmt.Println("Output:", stdout.String())
		fmt.Println("Error:", stderr.String())
		// if exitError, ok := err.(*exec.ExitError); ok {
		// 	if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
		// 		log.Fatal("Exit Code:", status.ExitStatus())
		// 	}
		// }
  }
	return stdout.String(), err
}

func SubmitTransaction () error {
	cmd := exec.Command("cardano-cli",
		"transaction",
		"submit",
		"--tx-file",
		"tx.signed",
		"--testnet-magic",
		testnetMagic,
	)
	_, err := execCommand(cmd)
	return err
}

var minLovelace = int64(2500000)

// Helper to build a valid transaction
func BuildTransaction( tx Transaction ) error {
	// The first build is for estimation, we pass in a wide enough lovelace value so the
	// size is approximated correctly.
	BuildTransactionRaw(tx, minLovelace)
	EnsureProtocolParameters()
	fee, err := EstimateFee(tx)
	if err != nil {
		return err
	}
	BalanceTransaction(&tx, fee)
	BuildTransactionRaw(tx, fee)
	return err
}

// Sign a transaction with the SPAddress.
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

	execCommand(cmd)
}

// Estimate fee of transaction
func EstimateFee( tx Transaction ) (fee int64, err error) {
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
		testnetMagic,
		)

	output, err := execCommand(cmd)
	if err != nil {
		return 0, err
	}

	_, fee_err := fmt.Sscan(output, &fee)
	if (fee_err != nil) {
		log.Fatal(fee_err)
	}

	return fee, err
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
		testnetMagic)

	execCommand(cmd)
}

// Get the transaction hash of the most recently built transaction.
func GetTransactionHash() (hash string, err error) {
	cmd := exec.Command("cardano-cli", "transaction", "txid", "--tx-file", "tx.signed")

	output, err := execCommand(cmd)

	hash = strings.Trim(output, "\n")
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

		if (output.DatumFile != "") {
			args = append(args, "--tx-out-inline-datum-file")
			args = append(args, output.DatumFile)
		}
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

func WriteRedeemerFile (path string, response *oracle.RewardResponse, redeemer Redeemer) {
	r, _ := regexp.Compile(`B \\"(.*?)\\"`)
	// NOTE: The submatch or capture group is the second argument, the first is the whole matched expression
	action_capture := r.FindStringSubmatch(response.Action)[1]
	datum_capture := r.FindStringSubmatch(response.Datum)[1]

	var action_hex = hex.EncodeToString([]byte(action_capture))
	var datum_hex = hex.EncodeToString([]byte(datum_capture))

	if err := os.WriteFile(path, []byte(fmt.Sprintf(
		REDEEMER_FORMAT_STRING,
		redeemer,
		response.SignatureDatum,
		response.MessageHashDatum,
		datum_hex,
		response.SignatureAction,
		response.MessageHashAction,
		action_hex)), 0666); err != nil {
		log.Fatal(err)
	}
}

func WriteDatumFile (path string, ntx int64, spPubKeyHash string, cpPubKeyHash string) {
	if err := os.WriteFile(path, []byte(fmt.Sprintf(DATUM_FORMAT_STRING, spPubKeyHash, cpPubKeyHash, ntx, PreGenMetaDataHash, PreGenWithdrawHash)), 0666); err != nil {
		log.Fatal(err)
	}
}
