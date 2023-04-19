package libp2p

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

const CIRendevousPoint string = "testing-nunet"

func TestBootstrap(t *testing.T) {
	ctx := context.Background()
	priv, _, _ := GenerateKey(0)
	host, idht, _ := NewHost(ctx, 9000, priv, true)
	defer host.Close()
	// Test successful Bootstrap
	err := Bootstrap(ctx, host, idht)
	if err != nil {
		t.Errorf("Expected Bootstrap to succeed but got error: %v", err)
	}

}

func TestSendDHTUpdate(t *testing.T) {
	// Setup nodes for testing
	ctx := context.Background()
	priv1, _, _ := GenerateKey(0)
	priv2, _, _ := GenerateKey(0)

	host1, idht1, err := NewHost(ctx, 9500, priv1, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host1.Close()
	host1.SetStreamHandler(DHTProtocolID, func(s network.Stream) {
		peerInfo := models.PeerData{}
		var peerID peer.ID
		data, err := io.ReadAll(s)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		err = json.Unmarshal(data, &peerInfo)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		// Update Peerstore
		peerID, err = peer.Decode(peerInfo.PeerID)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}

		host1.Peerstore().Put(peerID, "peer_info", peerInfo)
	})
	err = Bootstrap(ctx, host1, idht1)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	host2, idht2, err := NewHost(ctx, 9501, priv2, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host2.Close()
	host2.SetStreamHandler(DHTProtocolID, func(s network.Stream) {
		peerInfo := models.PeerData{}
		var peerID peer.ID
		data, err := io.ReadAll(s)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		err = json.Unmarshal(data, &peerInfo)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		// Update Peerstore
		peerID, err = peer.Decode(peerInfo.PeerID)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}

		host2.Peerstore().Put(peerID, "peer_info", peerInfo)
	})
	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	go Discover(ctx, host1, idht1, CIRendevousPoint)
	go Discover(ctx, host2, idht2, CIRendevousPoint)

	host1.Peerstore().AddAddrs(host2.ID(), host2.Addrs(), peerstore.PermanentAddrTTL)
	host1.Peerstore().AddPubKey(host2.ID(), host2.Peerstore().PubKey(host2.ID()))

	// Connect the two nodes
	if err := host1.Connect(ctx, host2.Peerstore().PeerInfo(host2.ID())); err != nil {
		t.Fatalf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	for host1.Network().Connectedness(host2.ID()).String() != "Connected" {
		if err := host1.Connect(ctx, host2.Peerstore().PeerInfo(host2.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	// Add mock data to their respective DHTs
	peerInfo1 := models.PeerData{}
	peerInfo1.PeerID = host1.ID().String()
	peerInfo1.HasGpu = true
	peerInfo1.AllowCardano = false
	host1.Peerstore().Put(host1.ID(), "peer_info", peerInfo1)
	peerInfo2 := models.PeerData{}
	peerInfo2.PeerID = host2.ID().String()
	peerInfo2.HasGpu = false
	peerInfo2.AllowCardano = true
	host2.Peerstore().Put(host2.ID(), "peer_info", peerInfo2)

	// Exchange DHT updates

	stream, err := host1.NewStream(ctx, host2.ID(), DHTProtocolID)
	if err != nil {
		t.Fatalf("Unable to open stream to host1: %v", err)
	}
	SendDHTUpdate(peerInfo1, stream)
	stream.Close()

	time.Sleep(100 * time.Millisecond)

	stream, err = host2.NewStream(ctx, host1.ID(), DHTProtocolID)
	if err != nil {
		t.Fatalf("Unable to open stream to host1: %v", err)
	}
	t.Log("Sending to 2--- ", host1.ID())
	SendDHTUpdate(peerInfo2, stream)
	stream.Close()

	time.Sleep(100 * time.Millisecond)

	machines1, err := host1.Peerstore().Get(host2.ID(), "peer_info")
	if err != nil {
		t.Fatalf("DHT update not received from Host2: %v", err)
	}
	machines2, err := host2.Peerstore().Get(host1.ID(), "peer_info")
	if err != nil {
		t.Fatalf("DHT update not received from Host1: %v", err)
	}

	assert.Equal(t, machines1, peerInfo2)
	assert.Equal(t, machines2, peerInfo1)

}

func TestFetchDhtContents(t *testing.T) {

	ctx := context.Background()

	priv1, _, _ := GenerateKey(0)
	priv2, _, _ := GenerateKey(0)

	host1, idht1, err := NewHost(ctx, 9500, priv1, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host1.Close()
	host1.SetStreamHandler(DHTProtocolID, func(s network.Stream) {
		peerInfo := models.PeerData{}
		var peerID peer.ID
		data, err := io.ReadAll(s)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		err = json.Unmarshal(data, &peerInfo)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		peerID, err = peer.Decode(peerInfo.PeerID)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		host1.Peerstore().Put(peerID, "peer_info", peerInfo)
	})
	err = Bootstrap(ctx, host1, idht1)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	host2, idht2, err := NewHost(ctx, 9501, priv2, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host2.Close()
	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	go Discover(ctx, host1, idht1, CIRendevousPoint)
	go Discover(ctx, host2, idht2, CIRendevousPoint)

	host2.Peerstore().AddAddrs(host1.ID(), host1.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(host1.ID(), host1.Peerstore().PubKey(host1.ID()))

	if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
		t.Fatalf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	for host2.Network().Connectedness(host1.ID()).String() != "Connected" {
		if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	peerInfo2 := models.PeerData{}
	peerInfo2.PeerID = host2.ID().String()
	peerInfo2.HasGpu = false
	host2.Peerstore().Put(host2.ID(), "peer_info", peerInfo2)

	stream, err := host2.NewStream(ctx, host1.ID(), DHTProtocolID)
	if err != nil {
		t.Fatalf("Unable to open stream to host1: %v", err)
	}
	SendDHTUpdate(peerInfo2, stream)
	stream.Close()

	time.Sleep(100 * time.Millisecond)

	contents := fetchDhtContents(host1)
	t.Log(contents)
	assert.Equal(t, peerInfo2, contents[0])
}

func TestFetchMachines(t *testing.T) {
	ctx := context.Background()

	priv1, _, _ := GenerateKey(0)
	priv2, _, _ := GenerateKey(0)

	host1, idht1, err := NewHost(ctx, 9500, priv1, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host1.Close()
	host1.SetStreamHandler(DHTProtocolID, func(s network.Stream) {
		peerID := s.Conn().RemotePeer()
		peerInfo := models.PeerData{}
		// var peerID peer.ID
		data, err := io.ReadAll(s)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		err = json.Unmarshal(data, &peerInfo)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		// peerID, err = peer.Decode(peerInfo.PeerID)
		// if err != nil {
		// 	t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		// }
		host1.Peerstore().Put(peerID, "peer_info", peerInfo)
	})
	err = Bootstrap(ctx, host1, idht1)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	host2, idht2, err := NewHost(ctx, 9501, priv2, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host2.Close()
	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	go Discover(ctx, host1, idht1, CIRendevousPoint)
	go Discover(ctx, host2, idht2, CIRendevousPoint)

	host2.Peerstore().AddAddrs(host1.ID(), host1.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(host1.ID(), host1.Peerstore().PubKey(host1.ID()))

	if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
		t.Fatalf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	for host2.Network().Connectedness(host1.ID()).String() != "Connected" {
		if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	peerInfo2 := models.PeerData{}
	peerInfo2.PeerID = host2.ID().String()
	peerInfo2.HasGpu = false
	host2.Peerstore().Put(host2.ID(), "peer_info", peerInfo2)

	stream, err := host2.NewStream(ctx, host1.ID(), DHTProtocolID)
	if err != nil {
		t.Fatalf("Unable to open stream to host1: %v", err)
	}
	SendDHTUpdate(peerInfo2, stream)
	stream.Close()

	time.Sleep(100 * time.Millisecond)

	machines := FetchMachines(host1)

	assert.Equal(t, peerInfo2, machines[host2.ID().String()])
}

func TestFetchAvailableResources(t *testing.T) {
	ctx := context.Background()

	priv1, _, _ := GenerateKey(0)
	priv2, _, _ := GenerateKey(0)

	host1, idht1, err := NewHost(ctx, 9500, priv1, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host1.Close()
	host1.SetStreamHandler(DHTProtocolID, func(s network.Stream) {
		peerInfo := models.PeerData{}
		var peerID peer.ID
		data, err := io.ReadAll(s)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		err = json.Unmarshal(data, &peerInfo)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		peerID, err = peer.Decode(peerInfo.PeerID)
		if err != nil {
			t.Fatalf("DHTUpdateHandler error: %s", err.Error())
		}
		host1.Peerstore().Put(peerID, "peer_info", peerInfo)
	})
	err = Bootstrap(ctx, host1, idht1)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	host2, idht2, err := NewHost(ctx, 9501, priv2, true)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	defer host2.Close()
	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}
	go Discover(ctx, host1, idht1, CIRendevousPoint)
	go Discover(ctx, host2, idht2, CIRendevousPoint)

	host2.Peerstore().AddAddrs(host1.ID(), host1.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(host1.ID(), host1.Peerstore().PubKey(host1.ID()))

	if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
		t.Fatalf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	for host2.Network().Connectedness(host1.ID()).String() != "Connected" {
		if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	peerInfo2 := models.PeerData{}
	peerInfo2.PeerID = host2.ID().String()
	peerInfo2.AvailableResources.Ram = 4000
	peerInfo2.AvailableResources.TotCpuHz = 14000

	host2.Peerstore().Put(host2.ID(), "peer_info", peerInfo2)

	stream, err := host2.NewStream(ctx, host1.ID(), DHTProtocolID)
	if err != nil {
		t.Fatalf("Unable to open stream to host1: %v", err)
	}
	SendDHTUpdate(peerInfo2, stream)
	stream.Close()

	time.Sleep(100 * time.Millisecond)

	availRes := FetchAvailableResources(host1)
	assert.Equal(t, peerInfo2.AvailableResources, availRes[0])

}

func TestGetPeers(t *testing.T) {
	ctx := context.Background()
	// Initialize host and dht objects
	priv, _, _ := GenerateKey(0)
	host, idht, _ := NewHost(ctx, 9000, priv, true)

	p2p = *DMSp2pInit(host, idht)

	defer host.Close()
	// Get the peers for the rendezvous string "nunet"
	_, err := p2p.GetPeers(ctx, "nunet")

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
