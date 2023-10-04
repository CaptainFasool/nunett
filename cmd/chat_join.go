package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

func init() {
}

var chatJoinCmd = &cobra.Command{
	Use:     "join",
	Short:   "Join open chat stream",
	Example: "nunet chat join <chatID>",
	Long:    "",
	Run: func(cmd *cobra.Command, args []string) {
		err := validateJoinChatInput(args)
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Printf("\nFor joining chats, check:\n\tnunet chat join --help\n")

			os.Exit(1)
		}

		query := "streamID=" + args[0]

		joinURL, err := utils.InternalAPIURL("ws", "/api/v1/peers/chat/join", query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		client := &Client{}
		err = client.Initialize(joinURL)
		if err != nil {
			fmt.Println("Failed to initialize client:", err)
			os.Exit(1)
		}
		defer client.Conn.Close()

		var wg sync.WaitGroup

		wg.Add(3)

		go func() {
			client.ReadMessages()
			wg.Done()
			client.stop <- true
		}()
		go func() {
			client.WriteMessages()
			wg.Done()
			client.stop <- true
		}()

		go func() {
			client.HandleInterruptsAndPings()
			wg.Done()
			client.stop <- true
		}()

		wg.Wait()
	},
}

func validateJoinChatInput(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no chat ID specified")
	} else if len(args) > 1 {
		return fmt.Errorf("unable to join multiple chats")
	} else {
		chatID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("argument is not a valid chat ID")
		}

		openChats, err := getIncomingChatList()
		if err != nil {
			return err
		}

		if chatID >= len(openChats) {
			return fmt.Errorf("no incoming stream match chat ID specified")
		}
	}

	return nil
}
