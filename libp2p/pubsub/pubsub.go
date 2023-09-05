package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gitlab.com/nunet/device-management-service/libp2p"

	libp2pPS "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type PubSubPeer struct{ *libp2pPS.PubSub }

type PsTopicSubscription struct {
	topic *libp2pPS.Topic
	sub   *libp2pPS.Subscription

	subscribeOnce sync.Once
	closeOnce     sync.Once
}

var pubsubHost *PubSubPeer
var once sync.Once

// NewGossipPubSub creates a new GossipSub instance with the given host or returns
// an existing one if it has been previously created.
func NewGossipPubSub(ctx context.Context, host host.Host) (*PubSubPeer, error) {
	if pubsubHost != nil {
		return pubsubHost, nil
	}

	once.Do(func() {
		gs, err := libp2pPS.NewGossipSub(ctx, host)
		if err != nil {
			zlog.Sugar().Errorf("Failed to create gossipsub: %v", err)
			return
		}
		pubsubHost = &PubSubPeer{gs}
	})

	return pubsubHost, nil
}

// JoinTopic joins the given topic and subscribes to the topic.
func (psHost *PubSubPeer) JoinTopic(topicName string) (*PsTopicSubscription, error) {
	tp, err := psHost.Join(topicName)
	if err != nil {
		return nil,
			fmt.Errorf("Failed to join topic %v, Error: %v", topicName, err)
	}

	sub, err := tp.Subscribe()
	if err != nil {
		return nil,
			fmt.Errorf("Failed to subscribe to topic %v, Error: %v", topicName, err)
	}

	return &PsTopicSubscription{
		topic: tp,
		sub:   sub,
	}, nil
}

// Publish publishes the given message to the topic.
func (ts *PsTopicSubscription) Publish(msg any) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Failed to marshal message, Error: %v", err)
	}

	zlog.Debug("Publishing")
	err = ts.topic.Publish(context.Background(), msgBytes)
	if err != nil {
		return fmt.Errorf("Failed to publish message, Error: %v", err)
	}
	zlog.Sugar().Debug("Published message")

	return nil
}

// Unsubscribe unsubscribes from the topic subscription.
func (ts *PsTopicSubscription) Unsubscribe() {
	ts.sub.Cancel()
}

func (ts *PsTopicSubscription) listenForMessages(ctx context.Context, msgCh chan *libp2pPS.Message) {
	for {
		zlog.Sugar().Debug("here")
		msg, err := ts.sub.Next(ctx)
		zlog.Sugar().Debug("finalyy")
		if err != nil {

			if err == context.Canceled ||
				err == libp2pPS.ErrSubscriptionCancelled {

				zlog.Sugar().Infof("Libp2p Pubsub topic %v done: %v", ts.topic.String(), err)
			} else {
				zlog.Sugar().Infof(
					"Unexpected error for libp2p pubsub topic %v done: %v", ts.topic.String(), err)
			}
			return
		}

		if msg.GetFrom().String() == libp2p.GetP2P().Host.ID().String() {
			continue
		}

		msgCh <- msg
		zlog.Sugar().Debugf("(%v): %v", msg.GetFrom().String(), msg.Message.String())

	}
}

func (ts *PsTopicSubscription) Close(ctx context.Context) error {
	var err error
	ts.closeOnce.Do(func() {
		zlog.Sugar().Infof("Closing subscription and topic itself for %v", ts.topic.String())
		if ts.sub != nil {
			ts.sub.Cancel()
		}
		if ts.topic != nil {
			err = ts.topic.Close()
		}

	})

	if err != nil {
		return err
	}

	return nil
}
