package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/utils"
)

func init() {

}

var chatListCmd = &cobra.Command{
	Use:   "list",
	Short: "Display table of open chat streams",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		body, err := utils.ResponseBody(nil, "GET", "/api/v1/peers/chat", nil)
		if err != nil {
			fmt.Println("Error trying to get response:", err)
			os.Exit(1)
		}

		if errMsg, err := jsonparser.GetString(body, "error"); err == nil { // if field "error" IS found
			fmt.Println("Error:", errMsg)
			os.Exit(1)
		} else if err == jsonparser.KeyPathNotFoundError { // if field "error" is NOT found; sucess
			var chatList []libp2p.OpenStream
			err := json.Unmarshal(body, &chatList)
			if err != nil {
				fmt.Println("Error trying to unmarshal", err)
				os.Exit(1)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Stream ID", "From Peer", "Time Opened"})
			table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
			table.SetCenterSeparator("|")
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			for _, chat := range chatList {
				table.Append([]string{strconv.Itoa(chat.ID), chat.StreamID, chat.FromPeer, chat.TimeOpened})
			}

			table.Render()
		} else { // if another error occurred
			fmt.Println("Error while parsing response:", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}
