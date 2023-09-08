package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	libp2pPS "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

// PubSub is basically a wrapper around *libp2p.PubSub original struct
// which has one additional method for joining and subscribing to a given topic
type PubSub struct {
	*libp2pPS.PubSub
	hostID            string
	joinSubscribeOnce sync.Once
}

// PsTopicSubscription is returned when the host joins/subs to a topic using the
// created JoinSubscribeTopic method. With its values, the host can deal with all
// the available attributes and methods for the given *libp2p.Topic and *libp2p.Subscribe.
type PsTopicSubscription struct {
	topic *libp2pPS.Topic
	sub   *libp2pPS.Subscription

	hostID string

	closeOnce sync.Once
}

var (
	pubsubHost   *PubSub
	onceGossipPS sync.Once
)

// NewGossipPubSub creates a new GossipSub instance with the given host or returns
// an existing one if it has been previously created.
func NewGossipPubSub(ctx context.Context, host host.Host) (*PubSub, error) {
	if pubsubHost != nil {
		return pubsubHost, nil
	}

	onceGossipPS.Do(func() {
		gs, err := libp2pPS.NewGossipSub(ctx, host)
		if err != nil {
			zlog.Sugar().Errorf("Failed to create gossipsub: %v", err)
			return
		}
		pubsubHost = &PubSub{
			gs,
			host.ID().String(),
			sync.Once{},
		}

	})

	return pubsubHost, nil
}

// JoinSubscribeTopic joins the given topic and subscribes to the topic.
func (ps *PubSub) JoinSubscribeTopic(topicName string) (*PsTopicSubscription, error) {
	// TODO: I'm not sure if having both methods called in the same function is a good thing.
	// We'll discover along the way (I did this because they seem to be highly coupled)
	var err error
	var topicSub *PsTopicSubscription

	ps.joinSubscribeOnce.Do(func() {
		tp, err := ps.Join(topicName)
		if err != nil {
			err = fmt.Errorf("Failed to join topic %v, Error: %v", topicName, err)
			return
		}

		sub, err := tp.Subscribe()
		if err != nil {
			err = fmt.Errorf("Failed to subscribe to topic %v, Error: %v", topicName, err)
			return
		}
		zlog.Sugar().Debugf("Joined and subscribe to topic: %v", topicName)

		topicSub = &PsTopicSubscription{
			topic:  tp,
			sub:    sub,
			hostID: ps.hostID,
		}
	})

	if err != nil {
		return nil, err
	}

	if topicSub == nil {
		return nil, fmt.Errorf("Topic already joined and/or subscribed!")
	}

	return topicSub, nil
}

// Publish publishes the given message to the objected topic
func (ts *PsTopicSubscription) Publish(msg any) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Failed to marshal message, Error: %v", err)
	}

	err = ts.topic.Publish(context.Background(), msgBytes)
	if err != nil {
		return fmt.Errorf("Failed to publish message, Error: %v", err)
	}
	zlog.Sugar().Debug("Published message successfully")

	return nil
}

// ListenForMessages receives a channel of type *libp2pPS.Message and it sends
// all the messages received for a given topic to this channel. Ignoring
// the messages send by the host
func (ts *PsTopicSubscription) ListenForMessages(ctx context.Context, msgCh chan *libp2pPS.Message) {
	for {
		zlog.Sugar().Debugf("Waiting for message for topic: %v", ts.topic.String())
		msg, err := ts.sub.Next(ctx)
		if err != nil {
			zlog.Sugar().Infof("Libp2p Pubsub topic %v done: %v", ts.topic.String(), err)
			return
		}

		if msg.GetFrom().String() == ts.hostID {
			continue
		}

		zlog.Sugar().Debugf("(%v): %v", msg.GetFrom().String(), msg.Message.Data)
		msgCh <- msg
	}
}

// Unsubscribe unsubscribes from the topic subscription.
func (ts *PsTopicSubscription) Unsubscribe() {
	ts.sub.Cancel()
}

// Close closes both subscription and topic for the given values assigned
// to ts *PsTopicSubscription
func (ts *PsTopicSubscription) Close(ctx context.Context) error {
	var err error
	ts.closeOnce.Do(func() {
		if ts.sub != nil {
			ts.sub.Cancel()
		}
		if ts.topic != nil {
			err = ts.topic.Close()
		}
		zlog.Sugar().Infof("Closed subscription and topic itself for topic: %v", ts.topic.String())
	})

	if err != nil {
		return err
	}

	return nil
}
