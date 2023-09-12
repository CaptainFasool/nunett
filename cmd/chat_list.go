package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/libp2p"
)

func init() {

}

var chatListCmd = &cobra.Command{
	Use:   "list",
	Short: "Display table of open chat streams",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		chatList, err := libp2p.IncomingChatRequests()
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
