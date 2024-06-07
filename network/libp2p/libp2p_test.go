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
		"no scheduler": {
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
			expErr: "scheduler is nil",
		},
		"success": {
			config: &models.Libp2pConfig{
				PrivateKey:              &crypto.Secp256k1PrivateKey{},
				BootstrapPeers:          []multiaddr.Multiaddr{},
				Rendezvous:              "nunet-randevouz",
				Server:                  false,
				Scheduler:               background_tasks.NewScheduler(1),
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
	pingResult, err := peer1.Ping(context.TODO(), peer2.Host.ID().String(), 100*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, pingResult.Success)
	zeroMicro, err := time.ParseDuration("0µs")
	assert.NoError(t, err)
	assert.Greater(t, pingResult.RTT, zeroMicro)

	// setup a new peer and advertise specs
	// peer3 will connect to peer2 in a ring setup.
	peer2p2pAddrs, err := peer2.GetMultiaddr()
	assert.NoError(t, err)
	peer3Config := setupPeerConfig(t, 65514, peer2p2pAddrs)
	peer3, err := New(peer3Config)
	assert.NoError(t, err)
	assert.NotNil(t, peer3)

	err = peer3.Init(context.TODO())
	assert.NoError(t, err)
	err = peer3.Start(context.TODO())
	assert.NoError(t, err)

	// advertise key
	err = peer1.Advertise(context.TODO(), "who_am_i", []byte(`{"peer":"peer1"}`))
	assert.NoError(t, err)
	err = peer2.Advertise(context.TODO(), "who_am_i", []byte(`{"peer":"peer2"}`))
	assert.NoError(t, err)
	err = peer3.Advertise(context.TODO(), "who_am_i", []byte(`{"peer":"peer3"}`))
	assert.NoError(t, err)

	time.Sleep(40 * time.Millisecond)

	// get the peers who have the who_am_i key
	advertisements, err := peer1.Advertisements(context.TODO(), "who_am_i")
	assert.NoError(t, err)
	assert.NotNil(t, advertisements)
	assert.Len(t, advertisements, 3)

	// check if all peers have returned the correct data
	for _, v := range advertisements {
		switch v.PeerId {
		case peer1.Host.ID().String():
			{
				assert.Equal(t, []byte(`{"peer":"peer1"}`), v.Data)
			}
		case peer2.Host.ID().String():
			{
				assert.Equal(t, []byte(`{"peer":"peer2"}`), v.Data)
			}
		case peer3.Host.ID().String():
			{
				assert.Equal(t, []byte(`{"peer":"peer3"}`), v.Data)
			}
		}
	}

	// peer3 unadvertises
	err = peer3.Unadvertise(context.TODO(), "who_am_i")
	assert.NoError(t, err)
	time.Sleep(40 * time.Millisecond)

	// get the values again, it should be 2 peers only
	advertisements, err = peer1.Advertisements(context.TODO(), "who_am_i")
	assert.NoError(t, err)
	assert.NotNil(t, advertisements)
	assert.Len(t, advertisements, 2)
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
