package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	// "github.com/libp2p/go-libp2p/core/ping"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

// UpdateConnections updates the database with the current connections.
func UpdateConnections(conns []network.Conn) {
	connMap := make(map[string]network.Conn)
	for _, conn := range conns {
		peerID := conn.RemotePeer().String()
		connMap[peerID] = conn
	}

	for _, conn := range conns {
		peerID := conn.RemotePeer().String()
		multiaddrs := conn.RemoteMultiaddr().String()

		connection := models.Connection{
			PeerID:     peerID,
			Multiaddrs: multiaddrs,
		}

		if result := db.DB.Where("peer_id = ?", peerID).Assign(models.Connection{}).FirstOrCreate(&connection); result.Error != nil {
			zlog.Sugar().Errorf("failed to update or insert connection for peer ID %s: %w", peerID, result.Error)
		}
	}

	var connections []models.Connection
	if result := db.DB.Find(&connections); result.Error != nil {
		zlog.Sugar().Errorf("failed to find connections: %w", result.Error)
	}
	for _, connection := range connections {
		if _, ok := connMap[connection.PeerID]; !ok {
			if err := RemoveConnection(connection); err != nil {
				zlog.Sugar().Errorf("failed to remove connection for peer ID %s: %w", connection.PeerID, err)
			}
		}
	}

}

func RemoveConnection(conn models.Connection) error {
	result := db.DB.Where("peer_id = ?", conn.PeerID).Find(&conn)
	if result.Error != nil {
		return result.Error
	}
	db.DB.Delete(&conn)
	return nil
}

func GetConnections() []models.Connection {
	var connections []models.Connection
	result := db.DB.Find(&connections)
	if result.Error != nil {
		zlog.Sugar().Errorf("Error while finding connections: %v", result.Error)
	}
	return connections
}

func PingHandler(s network.Stream) {
	if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
		zlog.Sugar().Info("Received Ping message")
	}
	reader := bufio.NewReader(s)
	writer := bufio.NewWriter(s)
	data, err := reader.ReadString('\n')

	if err != nil {
		zlog.Sugar().Errorf("failed to read string from stream: %v\n", err)
		return
	}

	// refuse replying to ping if already running a job
	if IsDepRespStreamOpen() {
		zlog.Sugar().Info("Refusing to reply to a ping because already running a job")
		return
	}

	// Echo the string back over the stream.
	_, err = writer.WriteString(data)
	if err != nil {
		zlog.Sugar().Errorf("failed to echo string back over stream: %v\n", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush writer: %v\n", err)
		return
	}

}

func PingPeer(ctx context.Context, h host.Host, target peer.ID) models.PingResult {
	var pingResult models.PingResult
	start := time.Now()

	// Create a new stream to the target peer using the ping protocol
	stream, err := h.NewStream(ctx, target, PingProtocolID)
	if err != nil {
		if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
			zlog.Sugar().Errorf("failed to create stream to peer %s: %w", target, err)
		}
		pingResult.Success = false
		return pingResult
	}
	r := bufio.NewReader(stream)
	w := bufio.NewWriter(stream)
	_, err = w.WriteString("ping" + "\n")
	if err != nil {
		if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
			zlog.Sugar().Errorf("failed to write ping message: %w", err)
		}

		pingResult.Success = false
		return pingResult
	}
	err = w.Flush()
	if err != nil {
		if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
			zlog.Sugar().Errorf("failed to flush ping message: %w", err)
		}
		pingResult.Success = false
		return pingResult
	}

	time.Sleep(1 * time.Second)

	pongMsg, err := r.ReadString('\n')
	if err != nil {
		if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
			zlog.Sugar().Errorf("failed to read pong message: %w", err)
		}
		pingResult.Success = false
		return pingResult
	}

	// Check if the pong message is the same as the ping message
	if pongMsg != "ping"+"\n" {
		if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
			zlog.Sugar().Errorf("unexpected pong message: %s", string(pongMsg))
		}
		pingResult.Success = false
		return pingResult
	}

	duration := time.Since(start)
	pingResult.Success = true
	pingResult.RTT = duration
	stream.Close()

	return pingResult
}

type PubSub struct {
	Host  host.Host
	Ps    *pubsub.PubSub
	Topic *pubsub.Topic
	Sub   *pubsub.Subscription
}

func PubSubInit(node host.Host) *PubSub {
	return &PubSub{Host: node}
}

var Pbsb PubSub

// NewPubSub creates a new PubSub instance.
func NewGossipPubSub(ctx context.Context, host host.Host) (*pubsub.PubSub, error) {
	Pbsb := *PubSubInit(host)
	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		zlog.Sugar().Errorf("Failed to create gossipsub: %v", err)
		return nil, err
	}
	Pbsb.Ps = ps
	return ps, nil
}

// JoinTopic joins the given topic, subscribes to the topic.
func (ps *PubSub) JoinTopic(topicName string) error {
	tp, err := ps.Ps.Join(topicName)
	if err != nil {
		zlog.Sugar().Errorf("Failed to join topic: %v", err)
		return err
	}
	sub, err := tp.Subscribe()
	if err != nil {
		zlog.Sugar().Errorf("Failed to subscribe to topic: %v", err)
		return err
	}
	ps.Topic = tp
	ps.Sub = sub
	return nil
}

// Publish publishes the given message to the topic.
func (ps *PubSub) Publish(msg any) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		zlog.Sugar().Errorf("Failed to marshal message: %v", err)
		return err
	}
	err = ps.Topic.Publish(context.Background(), msgBytes)
	if err != nil {
		zlog.Sugar().Errorf("Failed to publish message: %v", err)
		return err
	}

	return nil
}

// Unsubscribe unsubscribes from the topic subscription.
func (ps *PubSub) Unsubscribe() {
	ps.Sub.Cancel()
}

type blankValidator struct {
	P2p *P2P
}

func (bv blankValidator) Validate(key string, value []byte) error {
	// Check if the key has the correct namespace
	if !strings.HasPrefix(key, customNamespace) {
		return errors.New("invalid key namespace")
	}

	components := strings.Split(key, "/")
	key = components[len(components)-1]
	var dhtUpdate models.KadDHTMachineUpdate

	err := json.Unmarshal(value, &dhtUpdate)
	if err != nil {
		zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
		return err
	}

	// Extract data and signature fields
	data := dhtUpdate.Data
	var peerInfo models.PeerData
	err = json.Unmarshal(dhtUpdate.Data, &peerInfo)
	if err != nil {
		zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
		return err
	}

	signature := dhtUpdate.Signature
	remotePeerID, err := peer.Decode(key)
	if err != nil {
		zlog.Sugar().Errorf("Error decoding peerID: %v", err)
		return errors.New("error decoding peerID")
	}
	// Get the public key of the remote peer from the peerstore
	// remotePeerPublicKey :=
	// blankValidator.p2p.Host.Peerstore().PubKey(remotePeerID)
	remotePeerPublicKey := bv.P2p.Host.Peerstore().PubKey(remotePeerID)

	if remotePeerPublicKey == nil {

		return errors.New("public key for remote peer not found in peerstore")
	}
	verify, err := remotePeerPublicKey.Verify(data, signature)
	if err != nil {
		zlog.Sugar().Errorf("Error verifying signature: %v", err)
		return err
	}
	if !verify {
		zlog.Sugar().Info("Invalid signature")
		return errors.New("invalid signature")
	}

	if len(value) == 0 {
		return errors.New("value cannot be empty")
	}
	return nil
}
func (blankValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }
