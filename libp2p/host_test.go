package libp2p

import (
	"context"
	"testing"
)

func TestNewHost(t *testing.T) {
	ctx := context.Background()
	priv, _, _ := GenerateKey(0)

	host, dht, err := NewHost(ctx, priv, true)
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

func TestNewHost1(t *testing.T) {
	ctx := context.Background()
	priv, _, _ := GenerateKey(0)
	var p2p P2P
	p2p.NewHost(ctx, priv, true)
	if p2p.Host == nil {
		t.Fatalf("Host should not be nil")
	}
	if p2p.DHT == nil {
		t.Fatalf("DHT should not be nil")
	}
	defer p2p.Host.Close()
}
