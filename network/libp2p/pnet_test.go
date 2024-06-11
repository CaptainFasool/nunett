package libp2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// TestPrivateSwarm tests connectivity between nodes in the same private swarm and
// it checks if connection trials with outside nodes fail as it should.
//
// TODO: sometimes it takes several seconds to run this test (maximum 15s)
func TestPrivateSwarm(t *testing.T) {
	numNodes := 3
	nodes := createTestNetwork(t, numNodes, true)

	for _, nodeOuter := range nodes {
		for _, nodeInner := range nodes {
			t.Run("ping between peers within the same private swarm", func(t *testing.T) {
				t.Parallel()
				if nodeInner.Host.ID() == nodeOuter.Host.ID() {
					return
				}
				res, err := nodeOuter.Ping(context.TODO(), nodeInner.Host.ID().String(), 5*time.Second)
				assert.NoError(t, err)
				assert.Equal(t, true, res.Success)
			})
		}
	}

	bootstrapMultiAddr, err := nodes[0].GetMultiaddr()
	assert.NoError(t, err)

	config := setupPeerConfig(t, 64400, bootstrapMultiAddr, false)
	p, err := New(config, afero.NewMemMapFs())
	assert.NoError(t, err)

	err = p.Init(context.TODO())
	assert.NoError(t, err)

	err = p.Start(context.TODO())
	assert.NoError(t, err)

	// Ping seems to rely on peerIDs which under the hood are then used to search for peers multiaddresses
	// which might fail if the peer has no access to this info through DHT or local information (peerstore).
	// A more reliable way to test the security of a private network is to try to dial peers directly
	// using their multiaddresses.
	for _, node := range nodes {
		t.Run("Connection trial with peer outside the private swarm", func(t *testing.T) {
			t.Parallel()
			// trial: outer peer trying to connect with a peer within a pnet
			err := p.Host.Connect(context.TODO(), peer.AddrInfo{ID: node.Host.ID(), Addrs: node.Host.Addrs()})
			assert.Error(t, err)

			// trial: peer within a private network trying to establish a connection with an outer peer
			err = node.Host.Connect(context.TODO(), peer.AddrInfo{ID: p.Host.ID(), Addrs: p.Host.Addrs()})
			assert.Error(t, err)
		})
	}
}
