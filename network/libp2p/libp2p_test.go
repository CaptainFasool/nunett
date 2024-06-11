package libp2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/afero"
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
				PNet: models.PNetConfig{
					WithSwarmKey: false,
				},
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
				PNet: models.PNetConfig{
					WithSwarmKey: false,
				},
			},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			libp2p, err := New(tt.config, afero.NewMemMapFs())
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
	peer1Config := setupPeerConfig(t, 65512, []multiaddr.Multiaddr{}, false)
	peer1, err := New(peer1Config, afero.NewMemMapFs())
	assert.NoError(t, err)
	assert.NotNil(t, peer1)
	err = peer1.Init(context.TODO())
	assert.NoError(t, err)
	err = peer1.Start(context.TODO())
	assert.NoError(t, err)

	// peers of peer1 should be one (itself)
	assert.Equal(t, peer1.Host.Peerstore().Peers().Len(), 1)

	// setup peer2 to connect to peer 1
	peer1p2pAddrs, err := peer1.GetMultiaddr()
	assert.NoError(t, err)
	peer2Config := setupPeerConfig(t, 65513, peer1p2pAddrs, false)
	peer2, err := New(peer2Config, afero.NewMemMapFs())
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
	peer3Config := setupPeerConfig(t, 65514, peer2p2pAddrs, false)
	peer3, err := New(peer3Config, afero.NewMemMapFs())
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

	// test publish/subscribe
	// in this test gossipsub messages are not always delivered from peer1 to other peers
	// and this makes the test flaky. To solve the flakiness we use the peer one for
	// both publishing and subscribing.
	subscribeData := make(chan []byte)
	err = peer1.Subscribe(context.TODO(), "blocks", func(data []byte) {
		subscribeData <- data
		close(subscribeData)
	})
	assert.NoError(t, err)

	// calling one time publish doesnt gurantee that the message has been sent.
	// without this we might not get a message in the subscribe and we would get
	// flaky test which will fail
	go func() {
		for {
			err = peer1.Publish(context.TODO(), "blocks", []byte(`{"block":"1"}`))
		}
	}()

	receivedData := <-subscribeData
	assert.EqualValues(t, []byte(`{"block":"1"}`), receivedData)
}
