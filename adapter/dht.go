package adapter

import (
	"bytes"
	"context"
	"errors"
	"log"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func fetchDht() (string, error) {
	// Set up a connection to the server.
	address := "localhost:9998"
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewNunetAdapterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.GetDhtContent(ctx, &GetDhtParams{})

	if err != nil {
		return "", errors.New("could not get dht contents")
	}

	return r.GetDhtContents(), nil
}

func preProcessDht(contents []byte) (trimmed []byte) {
	trimmed = bytes.ReplaceAll(contents, []byte("'"), []byte("\""))
	return
}

func FetchDht() ([]byte, error) {
	content, err := fetchDht()
	if err != nil {
		return nil, errors.New("error getting dht content")
	}
	b := []byte(content)
	b = preProcessDht(b)

	return b, nil
}

func PeersNonBusy(peers []Peer) []Peer {
	// TODO: Not Implemented. Implementation is deferred when DHT/peers have data related to
	// resource onboarded vs resource already used. This will only happen when adapter is
	// re-written using Golang (libp2p).
	return peers
}

func PeersWithCardanoAllowed(peers []Peer) []Peer {
	var cardanoAllowedPeers []Peer

	for idx, peer := range peers {
		if peer.PeerID.AllowCardano == "true" {
			cardanoAllowedPeers = append(cardanoAllowedPeers, peer)
		}
		_ = idx
	}

	return cardanoAllowedPeers
}
