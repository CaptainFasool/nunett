package libp2p

import (
	"context"

	"github.com/gin-gonic/gin"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

var Node host.Host
var DHT *dht.IpfsDHT
var Ctx context.Context

func CheckOnboarding() {
	// Checks for saved metadata and create a new host
	var libp2pInfo models.Libp2pInfo
	result := db.DB.Where("id = ?", 1).Find(&libp2pInfo)
	if result.Error != nil {
		zlog.Sugar().Fatalf("Error Reading Database for Node Info: %s\n", result.Error.Error())
	}
	if libp2pInfo.PrivateKey != nil {
		// Recreate private key
		priv, err := crypto.UnmarshalPrivateKey(libp2pInfo.PrivateKey)
		if err != nil {
			zlog.Sugar().Fatalf("Error Reading Privatekey: %s\n", err.Error())
		}
		RunNode(priv)
	}
}

func RunNode(priv crypto.PrivKey) {
	ctx := context.Background()

	host, dht, err := NewHost(ctx, 9000, priv)
	if err != nil {
		zlog.Sugar().Fatalf("Error Creating Host: %s\n", err.Error())
	}
	err = Bootstrap(ctx, host, dht)
	if err != nil {
		zlog.Sugar().Fatalf("Error During Bootstrap: %s\n", err.Error())
	}
	go Discover(ctx, host, dht, "nunet")
	Node = host
	DHT = dht
	Ctx = ctx
}

// ListPeers  godoc
// @Summary      Return list of peers currently connected to
// @Description  Gets a list of peers the libp2p node can see within the network and return a list of peers
// @Tags         run
// @Produce      json
// @Success      200  {string}	string
// @Router       /peers [get]
func ListPeers(c *gin.Context) {

	peers, err := getPeers(Ctx, Node, DHT, "nunet")
	if err != nil {
		c.JSON(500, gin.H{"error": "can not fetch peers"})
		zlog.Sugar().Fatalf("Error Can Not Fetch Peers: %s\n", err.Error())
	}
	c.JSON(200, peers)

}
