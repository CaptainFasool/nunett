package cmd

import (
	"fmt"
	"os"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

var flagD, flagN bool

func init() {
	peerListCmd.Flags().BoolVarP(&flagD, "dht", "d", false, "list only DHT peers")
}

var peerListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Display list of peers in the network",
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

		dhtFlag, _ := cmd.Flags().GetBool("dht")

		if !dhtFlag {
			bootPeer, err := getBootstrapPeers()
			if err != nil {
				fmt.Println("Error getting Bootstrap peers:", err)
				os.Exit(1)
			}

			fmt.Printf("Bootstrap peers (%d)\n", len(bootPeer))
			for _, b := range bootPeer {
				fmt.Println(b)
			}

			fmt.Printf("\n")
		}

		dhtPeer, err := getDHTPeers()
		if err != nil {
			fmt.Println("Error getting DHT peers", err)
			os.Exit(1)
		}

		fmt.Printf("DHT peers (%d)\n", len(dhtPeer))
		for _, d := range dhtPeer {
			fmt.Println(d)
		}
	},
}

func getDHTPeers() ([]string, error) {
	var dhtSlice []string

	bodyDht, err := utils.ResponseBody(nil, "GET", "/api/v1/peers/dht", nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get response body: %v", err)
	}

	errMsg, err := jsonparser.GetString(bodyDht, "error")
	if err == nil {
		return nil, fmt.Errorf(errMsg)
	}
	msg, err := jsonparser.GetString(bodyDht, "message")
	if err == nil {
		return nil, fmt.Errorf(msg)
	}

	_, err = jsonparser.ArrayEach(bodyDht, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		dhtSlice = append(dhtSlice, string(value))
	})
	if err != nil {
		return nil, fmt.Errorf("cannot iterate over DHT peer list: %v", err)
	}

	return dhtSlice, nil
}

func getBootstrapPeers() ([]string, error) {
	var bootSlice []string

	bodyBoot, err := utils.ResponseBody(nil, "GET", "/api/v1/peers", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get response body: %v", err)
	}

	errMsg, err := jsonparser.GetString(bodyBoot, "error")
	if err == nil {
		return nil, fmt.Errorf(errMsg)
	}
	msg, err := jsonparser.GetString(bodyBoot, "message")
	if err == nil {
		return nil, fmt.Errorf(msg)

	}

	_, err = jsonparser.ArrayEach(bodyBoot, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		id, err := jsonparser.GetString(value, "ID")
		if err != nil {
			fmt.Println("Error getting Bootstrap peer ID string:", err)
			os.Exit(1)
		}

		bootSlice = append(bootSlice, id)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot iterate over Bootstrap peer list: %v", err)
	}

	return bootSlice, nil
}
