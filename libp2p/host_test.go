package libp2p

import (
	"context"
	"testing"
)

func TestNewHost(t *testing.T) {
	ctx := context.Background()
	port := 9000
	priv, _, _ := GenerateKey(0)

	host, dht, err := NewHost(ctx, port, priv)
	if err != nil {
		t.Fatalf("NewHost returned error: %v", err)
	}
	if host == nil {
		t.Fatalf("Host should not be nil")
	}
	if dht == nil {
		t.Fatalf("DHT should not be nil")
	}
	defer host.Close()
}
