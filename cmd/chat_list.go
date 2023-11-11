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
		chatList, err := getIncomingChatList()
		if err != nil {
			fmt.Println("Error:", err)
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
		os.Exit(0)
	},
}

func getIncomingChatList() ([]libp2p.OpenStream, error) {
	chatList, err := utils.ResponseBody(nil, "GET", "/api/v1/peers/chat", "", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get response body: %v", err)
	}

	errMsg, err := jsonparser.GetString(chatList, "error")
	if err == nil {
		return nil, fmt.Errorf(errMsg)
	}

	incomingChatList := []libp2p.OpenStream{}
	err = json.Unmarshal(chatList, &incomingChatList)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response body: %v", err)
	}

	return incomingChatList, nil
}
