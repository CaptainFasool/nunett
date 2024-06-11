package network

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/afero"
	commonproto "gitlab.com/nunet/device-management-service/proto/generated/v1/common"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/network/libp2p"
)

// Messenger defines the interface for sending messages.
type Messenger interface {
	// SendMessage sends a message to the given address.
	SendMessage(ctx context.Context, addrs []string, msg []byte) error
}

type Network interface {
	// Messenger embedded interface
	Messenger

	// Init initializes the network
	Init(context.Context) error

	// Start starts the network
	Start(context context.Context) error

	// Stat returns the network information
	Stat() models.NetworkStats

	// Ping pings the given address and returns the PingResult
	// default timeout is 5 seconds
	Ping(ctx context.Context, address string, timeout time.Duration) (models.PingResult, error)

	// Advertise advertises the given data with the given adId
	// such as advertising device capabilities on the DHT
	Advertise(ctx context.Context, key string, data []byte) error

	// Unadvertise stops advertising data corresponding to the given adId
	Unadvertise(ctx context.Context, key string) error

	// Advertisements returns the network advertisement
	Advertisements(ctx context.Context, key string) ([]*commonproto.Advertisement, error)

	// Publish publishes the given data to the given topic if the network
	// type allows publish/subscribe functionality such as gossipsub or nats
	Publish(ctx context.Context, topic string, data []byte) error
	// Subscribe subscribes to the given topic and calls the handler function
	// if the network type allows it simmilar to Publish()
	Subscribe(ctx context.Context, topic string, handler func(data []byte)) error

	// Stop stops the network including any existing advertisments and subscriptions
	Stop() error
}

type Message interface {
	// Dial dials a host with the given address
	Dial(address models.SpecConfig) error

	// Send sends the given data to the given address
	Send(data []byte) error

	// ReceiveHandler sets the handler function for received data
	ReceiveHandler(handler func(data []byte)) error

	// Close closes the connection
	Close() error

	// Info returns the information about the connection
	Info() models.MessageInfo
}

// NewNetwork returns a new network given the configuration.
func NewNetwork(netConfig *models.NetworkConfig, fs afero.Fs) (Network, error) {
	// TODO: probable additional params to receive: DB, FileSystem
	if netConfig == nil {
		return nil, errors.New("network configuration is nil")
	}
	switch netConfig.Type {
	case models.Libp2pNetwork:
		ln, err := libp2p.New(&netConfig.Libp2pConfig, fs)
		return ln, err
	case models.NATSNetwork:
		return nil, errors.New("not implemented")
	default:
		return nil, fmt.Errorf("unsupported network type: %s", netConfig.Type)
	}
}
