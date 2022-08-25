// Package networking deals with enabling and managing networking inside firecracker virtual machines.
//
// NuNet uses 172.20.0.0-172.20.255.255. Please note that 3rd octet is for particular application.
// If an application is using more than 1 VM, 4th octet will be used for assigning multiple VM in that network.
// For simplicity, following will be the mapping between tap devices on host and subnet range:
// tap0 -> 172.20.0.x
// tap1 -> 172.20.1.x
// tap2 -> 172.20.2.x
// tap3 -> 172.20.3.x
// And so on...
package networking

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// getActiveInterface returns active interface used to connect to internet.
func getActiveInterface() (string, error) {
	cmd := "ip route | grep default"
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", errors.New("you seem to be offline")
	}
	cmd = "ip route | grep default | awk '{print $5}'"
	out, _ := exec.Command("bash", "-c", cmd).Output()
	res := string(out)
	res = strings.Trim(res, "\n")
	return res, nil
}

// NextTapDevice queries the host machine for a tap devices name in a sequencial manner,
// if it does not exist, then it is returned to the caller. Example: tap0, tap1, tap2 etc.
// The first available is returned.
func NextTapDevice() string {
	counter := -1

	for {
		counter++
		deviceName := fmt.Sprintf("tap%d", counter)
		if !isTapInUse(deviceName) {
			return deviceName
		}

	}
}

// ConfigureTapByName takes in a tap device name, and configures subnet range and makes
// some changes to iptables tables and chains.
func ConfigureTapByName(tap string) error {
	subnetCidr := subnetCidrFromTap(tap)
	currentIface, _ := getActiveInterface()

	commandString := fmt.Sprintf("ip tuntap add %s mode tap", tap)
	_, err := exec.Command("bash", "-c", commandString).Output()
	if err != nil {
		return err
	}

	commandString = fmt.Sprintf("ip addr add %s dev %s", subnetCidr, tap)
	_, err = exec.Command("bash", "-c", commandString).Output()
	if err != nil {
		return err
	}

	commandString = fmt.Sprintf("ip link set %s up", tap)
	_, err = exec.Command("bash", "-c", commandString).Output()
	if err != nil {
		return err
	}

	commandString = "echo 1 > /proc/sys/net/ipv4/ip_forward"
	_, err = exec.Command("bash", "-c", commandString).Output()
	if err != nil {
		return err
	}

	commandString = fmt.Sprintf("iptables -t nat -C POSTROUTING -o %s -j MASQUERADE || iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", currentIface, currentIface)
	exec.Command("bash", "-c", commandString).Output()

	commandString = "iptables -C FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT || iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT"
	exec.Command("bash", "-c", commandString).Output()

	commandString = fmt.Sprintf("iptables -C FORWARD -i %s -o %s -j ACCEPT || iptables -A FORWARD -i %s -o %s -j ACCEPT", tap, currentIface, tap, currentIface)
	exec.Command("bash", "-c", commandString).Output()

	return nil
}

// isTapInUse tells if tap with same name is registered or not.
func isTapInUse(tap string) bool {
	_, err := exec.Command("ip", "link", "show", tap).Output()
	return err == nil
}

// subnetCidrFromTap takes in last number from the tap device name and take it as third octet in CIDR.
func subnetCidrFromTap(tap string) string {
	expression := regexp.MustCompile("[0-9]+$")
	tapId := expression.FindString(tap)

	return fmt.Sprintf("172.20.%s.1/24", tapId)
}
