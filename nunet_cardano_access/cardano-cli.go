package nunet_cardano_access

import (
	"context"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)



func RunCommand(command string) (*ResultOutput, error) {
	// Set up a connection to the server.
	address := "dev.nunet.io:9998"
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// log.Fatalf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := NewCardanoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.Run(ctx, &CommandInputs{
		Command: command,
	})

	if err != nil {
		return nil, err
	}

	return r, nil
}
