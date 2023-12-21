package backend

import gonet "github.com/shirou/gopsutil/net"

type Network struct{}

func (n *Network) GetConnections(kind string) ([]gonet.ConnectionStat, error) {
	return gonet.Connections(kind)
}
