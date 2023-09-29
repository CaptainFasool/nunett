package tunneling

import "github.com/songgao/water"

// TUN is a struct containing the fields necessary
// to configure a system TUN device. Access the
// internal TUN device through TUN.Iface
type TUN struct {
	Iface *water.Interface
	MTU   int
	Src   string
	Dst   string
}

// Apply configures the specified options for a TUN device.
func (t *TUN) Apply(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(t); err != nil {
			return err
		}
	}
	return nil
}

// Option defines a TUN device modifier option.
type Option func(tun *TUN) error

// Address sets the local address and subnet for an interface.
// On MacOS devices use this function to set the Src Address
// for an interface and use DestAddress to set the destination ip.
func Address(address string) Option {
	return func(tun *TUN) error {
		return tun.setAddress(address)
	}
}

// MTU sets the Maximum Transmission Unit size for an interface.
func MTU(mtu int) Option {
	return func(tun *TUN) error {
		return tun.setMTU(mtu)
	}
}

// DestAddress sets the destination address for a point-to-point interface.
// Only use this option on MacOS devices.
func DestAddress(address string) Option {
	return func(tun *TUN) error {
		return tun.setDestAddress(address)
	}
}
