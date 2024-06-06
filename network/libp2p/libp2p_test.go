package libp2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/internal/background_tasks"
	"gitlab.com/nunet/device-management-service/models"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config *models.Libp2pConfig
		expErr string
	}{
		"no config": {
			config: nil,
			expErr: "config is nil",
		},
		"success": {
			config: &models.Libp2pConfig{
				PrivateKey:              &crypto.Secp256k1PrivateKey{},
				BootstrapPeers:          []multiaddr.Multiaddr{},
				Rendezvous:              "nunet-randevouz",
				Server:                  false,
				Scheduler:               nil,
				CustomNamespace:         "/nunet-dht-1/",
				ListenAddress:           []string{"/ip4/localhost/tcp/10209"},
				PeerCountDiscoveryLimit: 40,
				PrivateNetwork:          false,
			},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			libp2p, err := New(tt.config)
			if tt.expErr != "" {
				assert.Nil(t, libp2p)
				assert.EqualError(t, err, tt.expErr)
			} else {
				assert.NotNil(t, libp2p)
			}
		})
	}
}

func TestLibp2pMethods(t *testing.T) {
	// setup peer1
	peer1Config := setupPeerConfig(t, 65512, []multiaddr.Multiaddr{})
	peer1, err := New(peer1Config)
	assert.NoError(t, err)
	assert.NotNil(t, peer1)
	err = peer1.Init(context.TODO())
	assert.NoError(t, err)
	err = peer1.Start(context.TODO())
	assert.NoError(t, err)

	// peers of peer1 should be one (itself)
	assert.Equal(t, peer1.Host.Peerstore().Peers().Len(), 1)
	fmt.Println("peer addresses", peer1.Host.Addrs())

	// setup peer2 to connect to peer 1
	peer1p2pAddrs, err := peer1.GetMultiaddr()
	assert.NoError(t, err)
	peer2Config := setupPeerConfig(t, 65513, peer1p2pAddrs)
	peer2, err := New(peer2Config)
	assert.NoError(t, err)
	assert.NotNil(t, peer2)

	err = peer2.Init(context.TODO())
	assert.NoError(t, err)
	err = peer2.Start(context.TODO())
	assert.NoError(t, err)

	// sleep until the inernals of the host get updated
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 2, peer2.Host.Peerstore().Peers().Len())
	assert.Equal(t, 2, peer1.Host.Peerstore().Peers().Len())

	// peer2 pings peer1
	pingResult, err := peer1.Ping(context.TODO(), peer1.Host.ID().String(), 100*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, pingResult.Success)
}

func setupPeerConfig(t *testing.T, libp2pPort int, bootstrapPeers []multiaddr.Multiaddr) *models.Libp2pConfig {
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
		PrivateNetwork:          false,
	}

}
