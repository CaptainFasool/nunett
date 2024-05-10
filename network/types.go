// Package network provides a simple network interface for the rest of the DMS.
package network

// Network defines an interface provided by DMS to be implemented by various
// network providers. This could be libp2p, or other p2p providers.
type Network interface {
	// Config sets the configuration for the network in
	Config() error

	// Init initializes the node with config specific in Config() phase.
	// tags: Start
	Init() error

	// EventRegister sets handlers to handle events such as change of local address
	EventRegister() error

	// Dial connects to a peer
	// tags: ConnectPeer
	Dial() error

	// Listen listens on a connection. Example could be a stream for libp2p connection.
	Listen() error

	// Status returns status of current host in regards to implementation.
	// Output must follow a generic struct.
	Status() bool

	// Tears down network interface.
	Stop() error
}

type VPN interface {
	// Start takes in an initial routing table and starts the VPN.
	Start() error

	// AddPeer is for adding the peers after the VPN has been started.
	// This should also update the routing table with the new peer.
	AddPeer() error

	// RemovePeer is oppisite of AddPeer. It should also update the routing table.
	RemovePeer() error

	// Stop tears down the VPN.
	Stop() error
}
