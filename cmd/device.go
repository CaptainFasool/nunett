package cmd

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

var deviceCmd = NewDeviceCmd(networkService)

func NewDeviceCmd(net backend.NetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "device",
		Short:             "device related operations",
		Long:              `manage onboarded device`,
		PersistentPreRunE: isDMSRunning(net),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(deviceStatusCmd)
	cmd.AddCommand(deviceSetCmd)
	return cmd
}

var deviceStatusCmd = NewDeviceStatusCmd(utilsService)
var deviceSetCmd = NewDeviceSetCmd(utilsService)

func NewDeviceStatusCmd(utilsService backend.Utility) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display current device status",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := checkOnboarded(utilsService)
			if err != nil {
				return err
			}

			body, err := utilsService.ResponseBody(nil, "GET", "/api/v1/device/status", "", nil)
			if err != nil {
				return fmt.Errorf("unable to get /device/status response body: %w", err)
			}

			online, err := jsonparser.GetBoolean(body, "device", "is_available")
			if err != nil {
				return fmt.Errorf("failed to get device status from json response: %w", err)
			}

			if online {
				fmt.Fprintln(cmd.OutOrStdout(), "Status: online")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Status: offline")
			}

			return nil
		},
	}
}

func NewDeviceSetCmd(utilsService backend.Utility) *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set device online status",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := checkOnboarded(utilsService)
			if err != nil {
				return err
			}

			if len(args) != 1 {
				return fmt.Errorf("invalid number of arguments")
			}

			var statusJson []byte

			if args[0] == "online" {
				statusJson = []byte(`{"is_available": true}`)
			} else if args[0] == "offline" {
				statusJson = []byte(`{"is_available": false}`)
			} else {
				return fmt.Errorf("invalid argument")
			}

			body, err := utilsService.ResponseBody(nil, "POST", "/api/v1/device/status", "", statusJson)
			if err != nil {
				return fmt.Errorf("could not get response body: %w", err)
			}

			errResponse, err := jsonparser.GetString(body, "error")
			if err != nil {
				return fmt.Errorf("failed to get error string: %w", err)
			}
			if errResponse != "" {
				return fmt.Errorf("failed to change device status: %s", errResponse)
			}

			_, err = jsonparser.GetBoolean(body, "device", "is_available")
			if err != nil {
				return fmt.Errorf("failed to get device status from response: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Device status successfully updated")
			return nil
		},
	}
}
