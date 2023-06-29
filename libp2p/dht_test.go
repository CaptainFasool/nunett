package libp2p

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

const CIRendevousPoint string = "testing-nunet"

func TestBootstrap(t *testing.T) {
	ctx := context.Background()
	priv, _, _ := GenerateKey(0)
	host, idht, _ := NewHost(ctx, priv, true)
	defer host.Close()
	// Test successful Bootstrap
	err := Bootstrap(ctx, host, idht)
	if err != nil {
		t.Errorf("Expected Bootstrap to succeed but got error: %v", err)
	}

}


func TestGetPeers(t *testing.T) {
	ctx := context.Background()
	// Initialize host and dht objects
	priv, _, _ := GenerateKey(0)
	host, idht, _ := NewHost(ctx, priv, true)

	p2p = *DMSp2pInit(host, idht)

	defer host.Close()
	// Get the peers for the rendezvous string "nunet"
	_, err := p2p.getPeers(ctx, "nunet")

	// Check if there is no error
	if err != nil {
		t.Fatalf("getPeers returned error: %v", err)
	}

}

func TestPeersWithCardanoAllowed(t *testing.T) {
	var peers []models.PeerData
	var peer1, peer2, peer3 models.PeerData
	peer1.AllowCardano = true
	peer2.AllowCardano = true
	peer3.AllowCardano = false
	peers = append(peers, peer1, peer2, peer3)
	res := PeersWithCardanoAllowed(peers)

	assert.Equal(t, 2, len(res))

}

func TestPeersWithGPU(t *testing.T) {
	var peers []models.PeerData
	var peer1, peer2, peer3 models.PeerData
	peer1.HasGpu = true
	peer2.HasGpu = true
	peer3.HasGpu = false
	peers = append(peers, peer1, peer2, peer3)
	res := PeersWithGPU(peers)

	assert.Equal(t, 2, len(res))

}

func TestPeersWithMatchingSpec(t *testing.T) {
	var depReq models.DeploymentRequest
	depReq.Constraints.CPU = 4000
	depReq.Constraints.RAM = 2000

	var peers []models.PeerData
	var peer1, peer2, peer3 models.PeerData
	peer1.AvailableResources.TotCpuHz = 5000
	peer1.AvailableResources.Ram = 4000
	peer2.AvailableResources.TotCpuHz = 8000
	peer2.AvailableResources.Ram = 1500
	peer3.AvailableResources.TotCpuHz = 2000
	peer3.AvailableResources.Ram = 3000

	peers = append(peers, peer1, peer2, peer3)

	res := PeersWithMatchingSpec(peers, depReq)
	assert.Equal(t, 1, len(res))
}
