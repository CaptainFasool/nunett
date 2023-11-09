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
)

type Subnet struct {
	ctx    context.Context
	cancel context.CancelFunc

	// iface is the tun device used to pass packets between
	// Hyprspace and the user's machine.
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

type Peer struct {
	ID   string
	Addr string
}

type CreateAndInviteParams struct {
	PeersIDs []string
}

func NewSubnet(ctx context.Context, cancel context.CancelFunc) *Subnet {
	return &Subnet{
		ctx:           ctx,
		cancel:        cancel,
		tunDev:        nil,
		routingTable:  make(map[peer.ID]string),
		activeStreams: make(map[string]network.Stream),
	}
}

// CreateSubnetAndInvite setup a routing table, assigning IP addresses to given peers,
// send a message of type (msgSubnetCreationInvite) to each peer with the routing
// table. The client must get the routing table and use it to join the subnet.
func (s *Subnet) CreateSubnetAndInvite(peersIDs []string) error {
	host := GetP2P().Host
	peersIDs = append(peersIDs, host.ID().String())

	var err error
	s.routingTable, err = setupRoutingTable(peersIDs)
	s.reversedRoutingTable = reverseRoutingTable(s.routingTable)
	if err != nil {
		return fmt.Errorf(
			"Couldn't setup routing table: %w", err)
	}

	for peerID, _ := range s.routingTable {
		err := s.invitePeerToSubnet(peerID)
		if err != nil {
			return fmt.Errorf(
				"Couldn't invite peer %s to subnet; Error: %w",
				peerID.String(), err)
		}
	}

	// Create TUN interface
	tunDev, err := tun.New(
		"dms-tun",
		tun.Address(s.routingTable[host.ID()]),
		tun.MTU(1420),
	)

	// Find and create connection with peers within Subnet
	decodedPeersIDs := utils.MakeListOfDictKeys(s.routingTable)
	go dialPeersContinuously(s.ctx, host,
		GetP2P().DHT, decodedPeersIDs)

	// Activate TUN interface to be ready to receive/send packets
	// tun.New() just created, it didn't make active.
	err = tunDev.Up()
	if err != nil {
		return fmt.Errorf("Couldn't activate TUN interface: %w", err)
	}

	// The following is responsible for SENDING packet to other peers
	// tunDev.Iface.Read() is reading all the packets coming to the TUN
	// interface, so if I do `ping 10.0.0.1`, the Iface.Read() will read
	// the packets that I'm trying to send to someone. As you can see,
	// `dst` will get the destination address before writing to the libp2p stream
	var packet = make([]byte, 1420)
	go func() {
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
				plen, err := tunDev.Iface.Read(packet)
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
	}()

	return nil
}

func (s *Subnet) invitePeerToSubnet(peerID peer.ID) error {
	zlog.Sugar().Debugf("Creating stream with peer %s to subnet",
		peerID.String())
	stream, err := GetP2P().Host.NewStream(s.ctx, peerID, SubnetProtocolID)
	if err != nil {
		return fmt.Errorf(
			"Couldn't create stream with peer %s, Error: %w",
			peerID, err)
	}

	zlog.Sugar().Debugf("Created stream with peer %s to subnet", peerID.String())
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
		subnetMsgJson, peerID.String())
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
	s.activeStreams[s.routingTable[peerID]] = stream
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
