package libp2p

import (
	"context"
	"testing"
)

func TestBootstrap(t *testing.T) {
	ctx := context.Background()
	priv, _, _ := GenerateKey(0)
	host, idht, _ := NewHost(ctx, 9000, priv)

	// Test successful Bootstrap
	err := Bootstrap(ctx, host, idht)
	if err != nil {
		t.Errorf("Expected Bootstrap to succeed but got error: %v", err)
	}

}
