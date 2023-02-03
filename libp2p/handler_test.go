package libp2p

import (
	"context"
	"testing"
)

func TestGetPeers(t *testing.T) {
	ctx := context.Background()
	// Initialize host and dht objects
	priv, _, _ := GenerateKey(0)
	host, idht, _ := NewHost(ctx, 9000, priv)

	// Get the peers for the rendezvous string "nunet"
	_, err := getPeers(ctx, host, idht, "nunet")

	// Check if there is no error
	if err != nil {
		t.Fatalf("getPeers returned error: %v", err)
	}

}
