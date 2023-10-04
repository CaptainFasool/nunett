package cmd

import (
	"fmt"
	"os"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

var addrs []string

var peerSelfCmd = &cobra.Command{
	Use:    "self",
	Short:  "Display self peer info",
	Long:   ``,
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
		onboarded, err := utils.IsOnboarded()
		if err != nil {
			fmt.Println("Error checking onboard status:", err)
			os.Exit(1)
		}

		if !onboarded {
			fmt.Println(`Looks like your machine is not onboarded...

For onboarding, check:
    nunet onboard --help`)
			os.Exit(1)
		}

		body, err := utils.ResponseBody(nil, "GET", "/api/v1/peers/self", nil)
		if err != nil {
			fmt.Println("Error getting response body:", err)
			os.Exit(1)
		}

		id, err := selfPeerID(body)
		if err != nil {
			fmt.Println("Error fetching self peer ID:", err)
			os.Exit(1)
		}

		addrsByte, err := selfPeerAddrs(body)
		if err != nil {
			fmt.Println("Error fetching self peer addresses:", err)
			os.Exit(1)
		}

		// iterate through array and append each value to slice
		jsonparser.ArrayEach(addrsByte, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			addrs = append(addrs, string(value))
		})

		fmt.Println("Host ID:", id)
		for index, addr := range addrs {
			if index+1 == len(addrs) { // if element is the last one
				fmt.Printf("%s\n", addr)
			} else {
				fmt.Printf("%s, ", addr)
			}
		}
	},
}

func selfPeerID(body []byte) (string, error) {
	id, err := jsonparser.GetString(body, "ID")
	if err != nil {
		return "", fmt.Errorf("unable to get ID string: %v", err)
	}

	return id, nil
}

func selfPeerAddrs(body []byte) (addrsByte []byte, err error) {
	addrsByte, dataType, _, err := jsonparser.Get(body, "Addrs")
	if err != nil {
		return nil, fmt.Errorf("unable to get addresses field: %v", err)
	}

	if dataType != jsonparser.Array {
		return nil, fmt.Errorf("invalid data type: expected addresses field is not an array")
	}

	return addrsByte, nil
}
