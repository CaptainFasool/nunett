package cmd

import (
	"fmt"
	"os"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

func init() {

}

var clearChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Clear open chat streams",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		body, err := utils.ResponseBody(nil, "GET", "/api/v1/peers/chat/clear", nil)
		if err != nil {
			fmt.Println("Error trying to get response:", err)
			os.Exit(1)
		}

		if errMsg, err := jsonparser.GetString(body, "error"); err == nil { // if field "error" IS found
			fmt.Println("Error:", errMsg)
			os.Exit(1)
		} else if err == jsonparser.KeyPathNotFoundError { // if field "error" is NOT found
			msg, _ := jsonparser.GetString(body, "message")
			fmt.Println(msg)
		} else { // if another error occurred
			fmt.Println("Error parsing response:", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}
