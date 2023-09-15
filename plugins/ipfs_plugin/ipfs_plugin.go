package ipfs_plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/libp2p/pubsub"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/plugins_management"

	libp2pPS "github.com/libp2p/go-libp2p-pubsub"
)

type IPFSPlugin struct {
	ctx    context.Context
	cancel context.CancelFunc

	info models.PluginInfo
	ts   *pubsub.PsTopicSubscription

	process *os.Process
	port    string
	addr    string
}

const (
	loopbackIP      = "127.0.0.1"
	pubsubTopicName = "cid_distribution"
)

var (
	ipfsPlugin   *IPFSPlugin
	muIPFSPlugin sync.Mutex
)

func NewIPFSPlugin() *IPFSPlugin {
	muIPFSPlugin.Lock()
	defer muIPFSPlugin.Unlock()

	if ipfsPlugin != nil {
		return ipfsPlugin
	}

	p := &IPFSPlugin{}
	p.ctx, p.cancel = context.WithCancel(context.Background())

	p.addr = loopbackIP
	p.port = "31001"

	i := models.PluginInfo{}
	i.Name = "ipfs-plugin"
	i.ResourcesUsage.TotCpuHz = 1000
	i.ResourcesUsage.Ram = 4000

	p.info = i
	return p
}

// Run deals with the startup of IPFS-Plugin through exec.Command()
// in which the default path for plugins is $dms-root/plugins/executables
func (p *IPFSPlugin) Run(pluginsManager *plugins_management.PluginsInfoChannels) {
	zlog.Sugar().Debug("Starting ", p.info.Name)
	executablePath := fmt.Sprintf("%v/%v", config.GetConfig().General.PluginsPath, p.info.Name)
	cmd := exec.Command(executablePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		p.cancel()
		pluginsManager.ErrCh <- fmt.Errorf(
			"Couldn't execute cmd.Start() for: %v, Error: %w", p.info.Name, err,
		)
		return
	}

	p.process = cmd.Process
	pluginsManager.SucceedStartup <- &p.info
	zlog.Sugar().Infof("Plugin %v started, path: %v, pid: %v", p.info.Name, cmd.Path, p.process.Pid)

	err = p.enterStoragePubSub(p.ctx)
	if err != nil {
		p.cancel()
		pluginsManager.ErrCh <- fmt.Errorf(
			"Couldn't enter on PubSub topic for CID distribution, Error: %w", err,
		)
		return

	}

	go p.ReadAndPinTopicCIDs(p.ctx)

	err = cmd.Wait()
	if err != nil {
		p.cancel()
		pluginsManager.ErrCh <- fmt.Errorf(
			"Plugin %v exited, Error: %w", p.info.Name, err,
		)
		return
	}

	return
}

// Stop stops the IPFS-Plugin process and return an error if any.
func (p *IPFSPlugin) Stop(pluginsManager *plugins_management.PluginsInfoChannels) error {
	if p.process == nil {
		return fmt.Errorf("There is no assigned process for plugin %v", p.info.Name)
	}

	// free resources before killing the process
	err := p.process.Release()
	if err != nil {
		pluginsManager.ErrCh <- fmt.Errorf("Unable to release process resources, Error: %w", err)
		return err
	}

	err = p.process.Kill()
	if err != nil {
		pluginsManager.ErrCh <- fmt.Errorf("Unable to kill ipfs-plugin process, Error: %w", err)
		return err
	}

	p.process = nil
	return nil
}

// IsRunning checks if a IPFS-Plugin process is running sending a Signal SIGUSR1 to it
func (p *IPFSPlugin) IsRunning(pluginsManager *plugins_management.PluginsInfoChannels) (bool, error) {
	err := p.process.Signal(syscall.SIGUSR1)
	if err != nil {
		fmt.Printf("Error signaling process: %s\n", err)
		return false, err
	}
	return true, nil
}

func (p *IPFSPlugin) enterStoragePubSub(ctx context.Context) error {
	ps := libp2p.GetP2P().PubSub
	// var routingDiscovery = drouting.NewRoutingDiscovery(dmsHost.DHT)
	// opt := libp2pPS.WithDiscovery(routingDiscovery)
	// ps, err := pubsub.NewGossipPubSub(ctx, dmsHost.Host, opt)
	// if err != nil {
	// 	return fmt.Errorf("Couldn't init GossipSub for IPFS-Plugin, Error: %w", err)
	// }

	ts, err := ps.JoinSubscribeTopic(ctx, pubsubTopicName, true)
	if err != nil {
		return fmt.Errorf("Couldn't enter on pubsub topic for IPFS-Plugin, Error: %w", err)
	}

	p.ts = ts

	return nil
}

func (p *IPFSPlugin) ReadAndPinTopicCIDs(ctx context.Context) error {
	msgCh := make(chan *libp2pPS.Message)
	go p.ts.ListenForMessages(ctx, msgCh)

	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()

	var m string
	for {
		select {
		case <-ctx.Done():
			close(msgCh)
			zlog.Sugar().Debug("Context is done. Stop listening to messages (CIDs) from PubSub", m)
		case msg := <-msgCh:
			err := json.Unmarshal(msg.Data, &m)
			if err != nil {
				close(msgCh)
				zlog.Sugar().Errorf("Couldn't unmarshal message from PubSub, Error: %v", err)
			}
			zlog.Sugar().Debugf("Sending CID %v to pin", m)

			// Send call to IPFS-PLugin to pin the data
			if m != "" {
				err = pinBasedOnCidRPC(m)
				if err != nil {
					close(msgCh)
					zlog.Sugar().Errorf("Error: %v", err)
				}
			}
		case <-tick.C:
			zlog.Sugar().Debug("Interval of IPFS-Plugin, no messages (CIDs) received")
		}

	}

	return nil
}

func (p *IPFSPlugin) debugTopic() {
	for {
		p.ts.Publish("DSNABDMNSABDMNSAVDMSAVDBSANC")
		zlog.Sugar().Debug(p.ts.Topic.ListPeers())
		time.Sleep(10 * time.Second)
	}
}
