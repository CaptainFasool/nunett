package libp2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/internal/background_tasks"
	"gitlab.com/nunet/device-management-service/models"
)

func TestCreateTestNet(t *testing.T) {
	fs := afero.NewMemMapFs()
	numNodes := 5
	nodes := createTestNet(t, fs, numNodes)

	// verify the number of nodes
	assert.Equal(t, 5, len(nodes))

	time.Sleep(5 * time.Second)

	// verify if all nodes have more than 2 connections
	for _, node := range nodes {
		assert.Greater(t, len(node.Host.Network().Peers()), 2)
	}
}

func createTestNet(t *testing.T, fs afero.Fs, n int) []*Libp2p {
	var peers []*Libp2p

	// initiating and configuring a single bootstrap node
	bootstrapConfig := setupLibp2pConfig(t, 45600, []multiaddr.Multiaddr{})
	bootstrapNode, err := New(bootstrapConfig, fs)
	assert.NoError(t, err)

	err = bootstrapNode.Init(context.TODO())
	assert.NoError(t, err)

	err = bootstrapNode.Start(context.TODO())
	assert.NoError(t, err)

	bootstrapMultiAddr, err := bootstrapNode.GetMultiaddr()
	assert.NoError(t, err)

	peers = append(peers, bootstrapNode)

	// create the remaining hosts
	for i := 1; i < n; i++ {
		config := setupLibp2pConfig(t, 45600+i, bootstrapMultiAddr)
		p, err := New(config, fs)
		if err != nil {
			t.Fatalf("Failed to create peer: %v", err)
		}
		err = p.Init(context.TODO())
		if err != nil {
			t.Fatalf("Failed to initialize peer: %v", err)
		}
		err = p.Start(context.TODO())
		if err != nil {
			t.Fatalf("Failed to start peer: %v", err)
		}
		peers = append(peers, p)
	}

	return peers
}

func setupLibp2pConfig(t *testing.T, libp2pPort int, bootstrapPeers []multiaddr.Multiaddr) *models.Libp2pConfig {
	priv, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	assert.NoError(t, err)
	return &models.Libp2pConfig{
		PrivateKey:              priv,
		BootstrapPeers:          bootstrapPeers,
		Rendezvous:              "nunet-randevouz",
		Server:                  false,
		Scheduler:               background_tasks.NewScheduler(10),
		CustomNamespace:         "/nunet-dht-1/",
		ListenAddress:           []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", libp2pPort)},
		PeerCountDiscoveryLimit: 40,
		PNet: models.PNetConfig{
			WithSwarmKey: false,
		},
	}
}
