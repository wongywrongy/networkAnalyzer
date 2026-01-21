package network

import (
	"net"
)

// InterfaceInfo describes a local network interface and its addresses.
type InterfaceInfo struct {
	Name      string
	MTU       int
	MAC       string
	Flags     []string
	Addresses []string
}

// ListInterfaces returns the available interfaces with IPv4/IPv6 addresses.
func ListInterfaces() ([]InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []InterfaceInfo
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		addrStrings := make([]string, 0, len(addrs))
		for _, a := range addrs {
			addrStrings = append(addrStrings, a.String())
		}

		flags := make([]string, 0, 4)
		if iface.Flags&net.FlagUp != 0 {
			flags = append(flags, "up")
		}
		if iface.Flags&net.FlagLoopback != 0 {
			flags = append(flags, "loopback")
		}
		if iface.Flags&net.FlagMulticast != 0 {
			flags = append(flags, "multicast")
		}
		if iface.Flags&net.FlagBroadcast != 0 {
			flags = append(flags, "broadcast")
		}

		result = append(result, InterfaceInfo{
			Name:      iface.Name,
			MTU:       iface.MTU,
			MAC:       iface.HardwareAddr.String(),
			Flags:     flags,
			Addresses: addrStrings,
		})
	}
	return result, nil
}

