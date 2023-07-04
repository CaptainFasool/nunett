package libp2p

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

const CIRendevousPoint string = "testing-nunet"

func TestBootstrap(t *testing.T) {
	ctx := context.Background()
	priv, _, _ := GenerateKey(0)
	var p2p P2P
	p2p.NewHost(ctx, priv, true)
	defer p2p.Host.Close()
	err := p2p.BootstrapNode(ctx)
	if err != nil {
		t.Errorf("Expected Bootstrap to succeed but got error: %v", err)
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

func TestFetchDhtContents(t *testing.T) {

	ctx := context.Background()

	priv1, _, _ := GenerateKey(0)
	priv2, _, _ := GenerateKey(0)

	var host1, host2 P2P

	host1.NewHost(ctx, priv1, true)
	defer host1.Host.Close()

	err := host1.BootstrapNode(ctx)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	host2.NewHost(ctx, priv2, true)
	defer host2.Host.Close()
	err = host2.BootstrapNode(ctx)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	go host1.StartDiscovery(ctx, CIRendevousPoint)
	go host2.StartDiscovery(ctx, CIRendevousPoint)
	host1.SetStreamHandler(PingProtocolID, PingHandler)
	host2.SetStreamHandler(PingProtocolID, PingHandler)

	host1.Host.Peerstore().AddAddrs(host2.Host.ID(), host2.Host.Addrs(), peerstore.PermanentAddrTTL)
	host1.Host.Peerstore().AddPubKey(host2.Host.ID(), host2.Host.Peerstore().PubKey(host2.Host.ID()))

	host2.Host.Peerstore().AddAddrs(host1.Host.ID(), host1.Host.Addrs(), peerstore.PermanentAddrTTL)
	host2.Host.Peerstore().AddPubKey(host1.Host.ID(), host1.Host.Peerstore().PubKey(host1.Host.ID()))

	t.Log("Host1 ID: ", host1.Host.ID().String())
	peerInfo2 := models.PeerData{}
	peerInfo2.PeerID = host2.Host.ID().String()
	peerInfo2.HasGpu = false
	bytes, err := json.Marshal(peerInfo2)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	t.Log("Host2 ID: ", host2.Host.ID().String())
	err = host2.AddToKadDHT(bytes, customNamespace)
	if err != nil {
		t.Fatalf("AddToKadDHT returned error: %v", err)
	}
	time.Sleep(3 * time.Second)
	host1.GetDHTUpdates(ctx)
	time.Sleep(3 * time.Second)
	contents := host1.fetchPeerStoreContents()
	t.Log(contents)
	assert.Equal(t, peerInfo2, contents[0])
}

func TestGetPeers(t *testing.T) {
	ctx := context.Background()
	// Initialize host and dht objects
	priv, _, _ := GenerateKey(0)
	var p2p P2P
	p2p.NewHost(ctx, priv, true)

	_, err := p2p.getPeers(ctx, "nunet")

	// Check if there is no error
	if err != nil {
		t.Fatalf("getPeers returned error: %v", err)
	}

}
