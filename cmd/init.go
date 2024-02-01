package cmd

import (
	"fmt"
	"os"

	"github.com/coreos/go-systemd/sdjournal"
	"gitlab.com/nunet/device-management-service/cmd/backend"
	// "gitlab.com/nunet/device-management-service/dms/config"
)

// var dmsPort = config.GetConfig().Rest.Port

var (
	networkService    = &backend.Network{}
	webSocketClient   = &backend.WebSocket{}
	fileSystemService = &backend.OS{}
	resourceService   = &backend.Resources{}
	p2pService        = &backend.P2P{}
	utilsService      = &backend.Utils{}
	walletService     = &backend.Wallet{}
	journalService    *backend.Journal
)

func init() {
	j, err := sdjournal.NewJournal()
	if err != nil {
		fmt.Printf("Error: could not initialize sdjournal: %v\n", err)
		os.Exit(1)
	}

	journalService = backend.SetRealJournal(j)
}
