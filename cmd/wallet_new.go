package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

var flagEth, flagAda bool

func init() {
	walletNewCmd.Flags().BoolVarP(&flagEth, "ethereum", "e", false, "create Ethereum wallet")
	walletNewCmd.Flags().BoolVarP(&flagAda, "cardano", "c", false, "create Cardano wallet")
}

var walletNewCmd = &cobra.Command{
	Use:    "new",
	Short:  "Create new wallet",
	Long:   ``,
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
		eth, _ := cmd.Flags().GetBool("ethereum")
		ada, _ := cmd.Flags().GetBool("cardano")

		var pair *models.BlockchainAddressPrivKey
		var err error

		// limit wallet creation to one at a time
		if ada && eth {
			fmt.Println(`Error: cannot create both wallets

For creating a wallet, check:
    nunet wallet new --help`)
			os.Exit(1)
		} else if ada {
			pair, err = onboarding.GetCardanoAddressAndMnemonic()
		} else if eth {
			pair, err = onboarding.GetEthereumAddressAndPrivateKey()
		} else {
			fmt.Println(`Error: no wallet flag specified

For creating a wallet, check:
    nunet wallet new --help`)
			os.Exit(1)
		}

		// share error handling for both addresses
		if err != nil {
			fmt.Println("Error creating wallet address:", err)
			os.Exit(1)
		}

		printWallet(pair)
		os.Exit(0)
	},
}

func printWallet(pair *models.BlockchainAddressPrivKey) {
	if pair.Address != "" {
		fmt.Printf("address: %s\n", pair.Address)
	}

	if pair.PrivateKey != "" {
		fmt.Printf("private_key: %s\n", pair.PrivateKey)
	}

	if pair.Mnemonic != "" {
		fmt.Printf("mnemonic: %s\n", pair.Mnemonic)
	}
}
