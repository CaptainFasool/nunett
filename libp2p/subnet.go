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

// TODOs:
// - Get responses for invites. With a new type of Message named "msgVPNInviteResponse"
// with the reason why it was refused

// - Rethink type instatiation and possible VPN interface design

// - invitePeerToVPN() function considers that the peer is already within the routingTable,
// that will not be always the case. Assign a new IP if the peer is not on the routing table yet
// (this will probably be framed as "addPeerToVPN()")

// - On the vpnStreamHandler side, we need to check if the invite is really coupled with a job
// for security reasons (Attack vector: fake invites. Right now, peers are accepting any invite
// with valid params)

// - Assign jobID to the vpnID

const (
	msgVPNCreationInvite = "VPNCreationInvite"
	defaultTunIfaceName  = "dms-tun"
)

type VPN struct {
	ctx    context.Context
	cancel context.CancelFunc

	// tunDev is the tunneling interface used as a bridge between sent/received
	// packets from host to dest. In practice, the real transport happens with
	// libp2p streams
	tunDev *tun.TUN

	// routingTable is the map of participant peers of the vpn
	// and their vpn addresses (key: peer.ID, value: <peerVpnIP>)
	routingTable vpnRouter

	// activeStreams is a map of active streams to a peer
	// (key: <peerVpnIP>, value: network.Stream)
	activeStreams map[string]network.Stream
}

type vpnRouter map[peer.ID]string
type reversedVPNRouter map[string]peer.ID

type vpnMessage struct {
	MsgType string
	Msg     string
}

type CreateAndInviteParams struct {
	PeersIDs []string
}

// NewVPN creates a new vpn, setting up a routing table, assigning IP addresses
// to all given peers. It also creates and activates the tunneling interface and
// make countiounsly connection with peers in the routing table.
func NewVPN(ctx context.Context, cancel context.CancelFunc, peersIDs []string) (
	*VPN, error) {

	host := GetP2P().Host
	peersIDs = append(peersIDs, host.ID().String())

	zlog.Sugar().Infof("Setting up routing table for peers %v", peersIDs)
	routingTable, err := setupRoutingTable(peersIDs)
	if err != nil {
		return nil, fmt.Errorf(
			"Couldn't setup routing table: %w", err)
	}
	zlog.Sugar().Debugf("Routing table: %v", routingTable)

	tunDev, err := createActivateTunIface(defaultTunIfaceName, routingTable)
	if err != nil {
		return nil, fmt.Errorf(
			"Couldn't create and activate the TUN interface: %w", err)
	}

	vpn = &VPN{
		ctx:           ctx,
		cancel:        cancel,
		tunDev:        tunDev,
		routingTable:  routingTable,
		activeStreams: make(map[string]network.Stream),
	}

	if len(routingTable) != 0 {
		zlog.Sugar().Debug("Dialing peers in the routing table")
		// Find and create connection with peers within VPN
		decodedPeersIDs := utils.MakeListOfDictKeys(routingTable)
		go dialPeersContinuously(ctx, host,
			GetP2P().DHT, decodedPeersIDs)
	}

	zlog.Sugar().Debug("Redirecting sent packets to destination")
	go vpn.redirectSentPacketsToDst()
	return vpn, nil
}

// JoinVPN joins an existing VPN given a routing table
// Note: this is the one which will be used when deploying jobs
func JoinVPN(ctx context.Context, cancel context.CancelFunc, routingTable vpnRouter) (
	*VPN, error) {
	if len(routingTable) == 0 {
		return nil, fmt.Errorf(
			"Can't join an empty vpn")
	}

	tunDev, err := createActivateTunIface(defaultTunIfaceName, routingTable)
	if err != nil {
		return nil, fmt.Errorf(
			"Couldn't create and activate the TUN interface: %w", err)
	}

	vpn := &VPN{
		ctx:           ctx,
		cancel:        cancel,
		tunDev:        tunDev,
		routingTable:  routingTable,
		activeStreams: make(map[string]network.Stream),
	}

	// Find and create connection with peers within VPN
	host := GetP2P().Host
	decodedPeersIDs := utils.MakeListOfDictKeys(routingTable)
	go dialPeersContinuously(ctx, host,
		GetP2P().DHT, decodedPeersIDs)

	go vpn.redirectSentPacketsToDst()
	return vpn, nil

}

// redirectSentPacketsToDst redirects packages sent to the host's tunneling interface
// to the destination through a libp2p stream
func (v *VPN) redirectSentPacketsToDst() {
	// The following is responsible for SENDING packet to other peers
	// tunDev.Iface.Read() is reading all the packets coming to the TUN
	// interface, so if I do `ping 10.0.0.1`, the Iface.Read() will read
	// the packets that I'm trying to send to someone. As you can see,
	// `dst` will get the destination address before writing to the libp2p stream
	var packet = make([]byte, 1420)
	host := GetP2P().Host

	for {
		select {
		case <-vpn.ctx.Done():
			zlog.Sugar().Error("Closing all vpn streams if any")
			for dst, stream := range v.activeStreams {
				stream.Close()
				delete(v.activeStreams, dst)
			}

			if err := v.tunDev.SetDownAndDelete(); err != nil {
				zlog.Sugar().Errorf(
					"Error closing and deleting TUN interface: %v", err)
			}

			return
		default:
			// ping 10.0.0.1
			// Read in a packet from the tun interface.
			plen, err := v.tunDev.Iface.Read(packet)
			if err != nil {
				zlog.Sugar().Errorf(
					"Error reading packet from TUN interface: %v", err)
				continue
			}

			// Check if there is anything at all within the packet
			if plen == 0 {
				continue
			}

			// Decode the packet's destination address
			dst := net.IPv4(packet[16], packet[17], packet[18], packet[19]).String()
			zlog.Sugar().Debugf("Send packet to destination peer: %v", dst)

			// Check if we already have an open connection to the destination peer.
			stream, ok := v.activeStreams[dst]
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
				delete(v.activeStreams, dst)
			}

			// Check if the destination of the packet is a known peer within
			// the routing table.
			reversedRoutingTable := reverseRoutingTable(v.routingTable)
			if peer, ok := reversedRoutingTable[dst]; ok {
				zlog.Sugar().Debugf(
					"Didn't have an active stream with peer %v, creating one", dst)
				stream, err = host.NewStream(v.ctx, peer, VPNProtocolID)
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

				// if successfully creation and sent of packets, add to active activeStreams
				// so that it can be reused latter
				v.activeStreams[dst] = stream
			}
		}
	}
}

// invitePeersToVPN sends a message of type (msgVPNCreationInvite) to each
// given peer
func (v *VPN) invitePeersToVPN(peersIDs []string) {
	for _, peerID := range peersIDs {

		if peerID == GetP2P().Host.ID().String() {
			continue
		}

		err := v.invitePeerToVPN(peerID)
		if err != nil {
			zlog.Sugar().Errorf(
				"Couldn't invite peer %s to vpn; Error: %v",
				peerID, err)
		}
	}
}

// invitePeerToVPN sends a message of type (msgVPNCreationInvite) to
// a given peer containing the routing table. The client must get the routing
// table and use it to join the vpn as it's the only necessary information
// to join the vpn in a secure manner
func (v *VPN) invitePeerToVPN(peerID string) error {
	if peerID == GetP2P().Host.ID().String() {
		return fmt.Errorf("Host peer can not invite itself to the vpn")
	}

	decodedPID, err := peer.Decode(peerID)
	if err != nil {
		return fmt.Errorf(
			"Couldn't decode input peerID '%s', Error: %w",
			peerID, err)
	}

	zlog.Sugar().Debugf("Creating stream with peer %s to vpn",
		peerID)
	stream, err := GetP2P().Host.NewStream(v.ctx, decodedPID, VPNProtocolID)
	if err != nil {
		return fmt.Errorf(
			"Couldn't create stream with peer %s, Error: %w",
			peerID, err)
	}

	zlog.Sugar().Debugf("Created stream with peer %s to vpn", peerID)
	routingTableJson, err := json.Marshal(v.routingTable)
	if err != nil {
		stream.Close()
		return fmt.Errorf("failed to marshal routing table, Error: %w", err)
	}

	vpnMsg := vpnMessage{
		MsgType: msgVPNCreationInvite,
		Msg:     string(routingTableJson),
	}

	vpnMsgJson, err := json.Marshal(vpnMsg)
	if err != nil {
		stream.Close()
		return fmt.Errorf("failed to marshal vpn message, Error: %w", err)
	}

	zlog.Sugar().Debugf("Sending message %s to peer %s",
		vpnMsgJson, peerID)
	w := bufio.NewWriter(stream)
	_, err = w.WriteString(fmt.Sprintf("%s\n", vpnMsgJson))
	if err != nil {
		stream.Close()
		return fmt.Errorf(
			"Couldn't write message: %v to stream. Error: %w",
			msgVPNCreationInvite, err)
	}
	err = w.Flush()
	if err != nil {
		stream.Close()
		return fmt.Errorf(
			"Couldn't flush message: %v to stream. Error: %w",
			msgVPNCreationInvite, err)
	}
	v.activeStreams[v.routingTable[decodedPID]] = stream
	return nil
}

func setupRoutingTable(peersIDs []string) (vpnRouter, error) {
	var routingTable = make(map[peer.ID]string)

	for idx, peerID := range peersIDs {
		decodedPID, err := peer.Decode(peerID)
		if err != nil {
			return nil, fmt.Errorf(
				"Couldn't decode input peerID '%s', Error: %w",
				peerID, err)
		}

		routingTable[decodedPID] = fmt.Sprintf("10.0.0.%d/24", idx)
	}
	return routingTable, nil
}

func reverseRoutingTable(routingTable vpnRouter) reversedVPNRouter {
	var reversedRoutingTable = make(reversedVPNRouter)
	for peerID, vpnAddr := range routingTable {
		reversedRoutingTable[vpnAddr] = peerID
	}
	return reversedRoutingTable
}

func createActivateTunIface(tunName string, routingTable vpnRouter) (*tun.TUN, error) {
	zlog.Sugar().Debug("Creating TUN interface")
	// Create TUN interface
	tunDev, err := tun.New(
		tunName,
		tun.Address(routingTable[GetP2P().Host.ID()]),
		tun.MTU(1420),
	)
	if err != nil {
		return nil, fmt.Errorf("Couldn't create TUN interface: %w", err)
	}
	zlog.Sugar().Debug("TUN interface created")

	zlog.Sugar().Debug("Activating TUN interface")
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
	zlog.Sugar().Debug("Tunneling interface created and up")

	return tunDev, nil
}
