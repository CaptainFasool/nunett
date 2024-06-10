package libp2p

import (
	"context"
	"testing"
	"time"

	// "github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestWithSwarmKey(t *testing.T) {
	numNodes := 5
	nodes := createTestNetwork(t, numNodes, true)

	// test if peers within the private network can ping each other
	for _, nodeOuter := range nodes {
		for _, nodeInner := range nodes {
			res, err := nodeOuter.Ping(context.TODO(), nodeInner.Host.ID().String(), 5*time.Second)
			assert.NoError(t, err)
			assert.Equal(t, true, res.Success)
		}
	}

	// TODO: instantiate a new peer without the swarm key and see if it
	// can ping peers within the private network
	// 	bootstrapMultiAddr, err := nodes[0].GetMultiaddr()
	// 	assert.NoError(t, err)
	//
	// 	config := setupLibp2pConfig(t, 64400, bootstrapMultiAddr, false)
	// 	p, err := New(config, afero.NewMemMapFs())
	// 	assert.NoError(t, err)
	//
	// 	err = p.Init(context.TODO())
	// 	assert.NoError(t, err)
	//
	// 	err = p.Start(context.TODO())
	// 	assert.NoError(t, err)
}
