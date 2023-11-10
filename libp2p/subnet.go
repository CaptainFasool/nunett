package libp2p

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"

	tun "gitlab.com/nunet/device-management-service/network/tunneling"
	"gitlab.com/nunet/device-management-service/utils"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	msgSubnetCreationInvite = "SubnetCreationInvite"
	defaultTunIfaceName     = "dms-tun"
)

type Subnet struct {
	ctx    context.Context
	cancel context.CancelFunc

	// tunDev is the tunneling interface used as a bridge between sent/received
	// packets from host to dest. In practice, the real transport happens with
	// libp2p streams
	tunDev *tun.TUN

	// routingTable is the map of participant peers of the subnet
	// and their subnet addresses (key: peer.ID, value: <peerSubnetIP>)
	routingTable subnetTable

	// reverseRoutingTable being key: <peerSubnetIP>
	reversedRoutingTable reversedSubnetTable

	// activeStreams is a map of active streams to a peer
	// (key: <peerSubnetIP>, value: network.Stream)
	activeStreams map[string]network.Stream
}

type subnetTable map[peer.ID]string
type reversedSubnetTable map[string]peer.ID

type subnetMessage struct {
	MsgType string
	Msg     string
}

type CreateAndInviteParams struct {
	PeersIDs []string
}

// NewSubnet creates a new subnet, setting up a routing table, assigning IP addresses
// to all given peers. It also creates and activates the tunneling interface and
// make countiounsly connection with peers in the routing table.
func NewSubnet(ctx context.Context, cancel context.CancelFunc, peersIDs []string) (
	*Subnet, error) {

	host := GetP2P().Host
	peersIDs = append(peersIDs, host.ID().String())

	routingTable, err := setupRoutingTable(peersIDs)
	if err != nil {
		return nil, fmt.Errorf(
			"Couldn't setup routing table: %w", err)
	}
	reversedRoutingTable := reverseRoutingTable(routingTable)

	tunDev, err := createActivateTunIface(defaultTunIfaceName, routingTable)
	if err != nil {
		return nil, fmt.Errorf(
			"Couldn't create and activate the TUN interface: %w", err)
	}

	subnet := &Subnet{
		ctx:                  ctx,
		cancel:               cancel,
		tunDev:               tunDev,
		routingTable:         routingTable,
		reversedRoutingTable: reversedRoutingTable,
		activeStreams:        make(map[string]network.Stream),
	}

	if len(routingTable) != 0 {
		// Find and create connection with peers within Subnet
		decodedPeersIDs := utils.MakeListOfDictKeys(routingTable)
		go dialPeersContinuously(ctx, host,
			GetP2P().DHT, decodedPeersIDs)
	}

	go subnet.redirectSentPacketsToDst()
	return subnet, nil
}

// JoinSubnet joins an existing subnet given a routing table
func JoinSubnet(ctx context.Context, cancel context.CancelFunc, routingTable subnetTable) (
	*Subnet, error) {
	if len(routingTable) == 0 {
		return nil, fmt.Errorf(
			"Can't join an empty subnet")
	}

	tunDev, err := createActivateTunIface(defaultTunIfaceName, routingTable)
	if err != nil {
		return nil, fmt.Errorf(
			"Couldn't create and activate the TUN interface: %w", err)
	}

	subnet := &Subnet{
		ctx:                  ctx,
		cancel:               cancel,
		tunDev:               tunDev,
		routingTable:         routingTable,
		reversedRoutingTable: reverseRoutingTable(routingTable),
		activeStreams:        make(map[string]network.Stream),
	}

	host := GetP2P().Host
	if len(routingTable) != 0 {
		// Find and create connection with peers within Subnet
		decodedPeersIDs := utils.MakeListOfDictKeys(routingTable)
		go dialPeersContinuously(ctx, host,
			GetP2P().DHT, decodedPeersIDs)
	}

	go subnet.redirectSentPacketsToDst()
	return subnet, nil

}

// redirectSentPacketsToDst redirects packages sent to the host's tunneling interface
// to the destination through a libp2p stream
func (s *Subnet) redirectSentPacketsToDst() {
	// The following is responsible for SENDING packet to other peers
	// tunDev.Iface.Read() is reading all the packets coming to the TUN
	// interface, so if I do `ping 10.0.0.1`, the Iface.Read() will read
	// the packets that I'm trying to send to someone. As you can see,
	// `dst` will get the destination address before writing to the libp2p stream
	var packet = make([]byte, 1420)
	host := GetP2P().Host

	for {
		select {
		case <-subnet.ctx.Done():
			zlog.Sugar().Error("Closing all subnet streams if any")
			for dst, stream := range s.activeStreams {
				stream.Close()
				delete(s.activeStreams, dst)
			}

			if err := s.tunDev.SetDownAndDelete(); err != nil {
				zlog.Sugar().Errorf(
					"Error closing and deleting TUN interface: %v", err)
			}

			return
		default:
			// ping 10.0.0.1
			// Read in a packet from the tun interface.
			plen, err := s.tunDev.Iface.Read(packet)
			if err != nil {
				zlog.Sugar().Errorf(
					"Error reading packet from TUN interface: %v", err)
				continue
			}
			// TODO: check if there is anything at all within the packet

			// Decode the packet's destination address
			dst := net.IPv4(packet[16], packet[17], packet[18], packet[19]).String()
			zlog.Sugar().Debugf("Send packet to destination peer: %v", dst)

			// Check if we already have an open connection to the destination peer.
			stream, ok := s.activeStreams[dst]
			if ok {
				// Write out the packet's length to the libp2p stream to ensure
				// we know the full size of the packet at the other end.
				err = binary.Write(stream, binary.LittleEndian, uint16(plen))
				if err == nil {
					// Write the packet out to the libp2p stream.
					// If everyting succeeds continue on to the next packet.
					_, err = stream.Write(packet[:plen])
					if err == nil {
						zlog.Sugar().Debugf(
							"Successfully sent packet to: %v", dst)
						continue
					}
				}
				// If we encounter an error when writing to a stream we should
				// close that stream and delete it from the active stream map.
				zlog.Sugar().Errorf(
					"Error writing to libp2p stream from tunneling: %v", err)
				stream.Close()
				delete(s.activeStreams, dst)
			}

			// Check if the destination of the packet is a known peer to
			// the interface.
			if peer, ok := s.reversedRoutingTable[dst]; ok {
				zlog.Sugar().Debugf(
					"Didn't have an active stream with peer %v, creating one", dst)
				stream, err = host.NewStream(s.ctx, peer, SubnetProtocolID)
				if err != nil {
					zlog.Sugar().Errorf(
						"Error creating stream with peer: %v", dst)
					continue
				}
				// Write packet length
				err = binary.Write(stream, binary.LittleEndian, uint16(plen))
				if err != nil {
					zlog.Sugar().Error("Error writing packet size")
					stream.Close()
					continue
				}
				// Write the packet
				_, err = stream.Write(packet[:plen])
				if err != nil {
					zlog.Sugar().Errorf(
						"Error writing to libp2p stream from tunneling: %v", err)
					stream.Close()
					continue
				}
				zlog.Sugar().Debugf(
					"Successfully sent packet to: %v", dst)

				// If all succeeds when writing the packet to the stream
				// we should reuse this stream by adding it active streams map.
				s.activeStreams[dst] = stream
			}
		}
	}
}

// CreateSubnetAndInvite setup a routing table, assigning IP addresses to given peers,
// send a message of type (msgSubnetCreationInvite) to each peer with the routing
// table. The client must get the routing table and use it to join the subnet.

// invitePeersToSubnet sends a message of type (msgSubnetCreationInvite) to each
// given peer
func (s *Subnet) invitePeersToSubnet(peersIDs []string) {
	for _, peerID := range peersIDs {

		if peerID == GetP2P().Host.ID().String() {
			continue
		}

		err := s.invitePeerToSubnet(peerID)
		if err != nil {
			zlog.Sugar().Errorf(
				"Couldn't invite peer %s to subnet; Error: %w",
				peerID, err)
		}
	}
}

// invitePeerToSubnet sends a message of type (msgSubnetCreationInvite) to
// a given peer containing the routing table. The client must get the routing
// table and use it to join the subnet as it's the only necessary information
// to join the subnet in a secure manner
func (s *Subnet) invitePeerToSubnet(peerID string) error {
	if peerID == GetP2P().Host.ID().String() {
		return fmt.Errorf("Host peer can not invite itself to the subnet")
	}

	decodedPID, err := peer.Decode(peerID)
	if err != nil {
		return fmt.Errorf(
			"Couldn't decode input peerID '%s', Error: %w",
			peerID, err)
	}

	zlog.Sugar().Debugf("Creating stream with peer %s to subnet",
		peerID)
	stream, err := GetP2P().Host.NewStream(s.ctx, decodedPID, SubnetProtocolID)
	if err != nil {
		return fmt.Errorf(
			"Couldn't create stream with peer %s, Error: %w",
			peerID, err)
	}

	zlog.Sugar().Debugf("Created stream with peer %s to subnet", peerID)
	routingTableJson, err := json.Marshal(s.routingTable)
	if err != nil {
		stream.Close()
		return fmt.Errorf("failed to marshal routing table, Error: %w", err)
	}

	subnetMsg := subnetMessage{
		MsgType: msgSubnetCreationInvite,
		Msg:     string(routingTableJson),
	}

	subnetMsgJson, err := json.Marshal(subnetMsg)
	if err != nil {
		stream.Close()
		return fmt.Errorf("failed to marshal subnet message, Error: %w", err)
	}

	zlog.Sugar().Debugf("Sending message %s to peer %s",
		subnetMsgJson, peerID)
	w := bufio.NewWriter(stream)
	_, err = w.WriteString(fmt.Sprintf("%s\n", subnetMsgJson))
	if err != nil {
		stream.Close()
		return fmt.Errorf(
			"Couldn't write message: %v to stream. Error: %w",
			msgSubnetCreationInvite, err)
	}
	err = w.Flush()
	if err != nil {
		stream.Close()
		return fmt.Errorf(
			"Couldn't flush message: %v to stream. Error: %w",
			msgSubnetCreationInvite, err)
	}
	s.activeStreams[s.routingTable[decodedPID]] = stream
	return nil
}

func setupRoutingTable(peersIDs []string) (subnetTable, error) {
	var routingTable = make(map[peer.ID]string)

	for idx, peerID := range peersIDs {
		decodedPID, err := peer.Decode(peerID)
		if err != nil {
			return nil, fmt.Errorf(
				"Couldn't decode input peerID '%s', Error: %w",
				peerID, err)
		}

		routingTable[decodedPID] = fmt.Sprintf("10.0.0.%d", idx)
	}
	return routingTable, nil
}

func reverseRoutingTable(routingTable subnetTable) reversedSubnetTable {
	var reversedRoutingTable = make(reversedSubnetTable)
	for peerID, subnetAddr := range routingTable {
		reversedRoutingTable[subnetAddr] = peerID
	}
	return reversedRoutingTable
}

func createActivateTunIface(tunName string, routingTable subnetTable) (*tun.TUN, error) {
	// Create TUN interface
	tunDev, err := tun.New(
		tunName,
		tun.Address(routingTable[GetP2P().Host.ID()]),
		tun.MTU(1420),
	)
	if err != nil {
		return nil, fmt.Errorf("Couldn't create TUN interface: %w", err)
	}

	// Activate TUN interface to be ready to receive/send packets
	// tun.New() just created, it didn't make active.
	err = tunDev.Up()
	if err != nil {
		if err := tunDev.Delete(); err != nil {
			zlog.Sugar().Errorf(
				"Error deleting TUN interface: %v", err)
		}
		return nil, fmt.Errorf("Couldn't activate TUN interface: %w", err)
	}

	return tunDev, nil
}
