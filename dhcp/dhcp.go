package dhcp

import (
	"errors"
	"math"
	"net"
)

type DHCPServer struct {
	Size           int
	NetworkAddress net.IP
	Current        int
	Used           map[string]bool
	Max            int
}

func NewDHCPServer(cidr string) (*DHCPServer, error) {
	ipv4Start, ipv4Net, err := net.ParseCIDR(cidr)

	if err != nil {
		return nil, err
	}

	size, _ := ipv4Net.Mask.Size()

	if size > 30 {
		return nil, errors.New("Subnet is too small, CIDR should be at least /30")
	}

	max := int(math.Pow(2.0, 32.0-float64(size))) - 1 // Skip the broadcast address

	return &DHCPServer{
		Size:           size,
		Max:            max,
		NetworkAddress: ipv4Net.IP,
		Current:        int(ipv4Start[len(ipv4Start)-1]),
		Used:           make(map[string]bool),
	}, nil
}

func (dhcp *DHCPServer) Pop() (string, error) {
	for i := 0; i < dhcp.Max; i++ {
		next := dhcp.NextIP()
		if !dhcp.Used[next] {
			dhcp.Used[next] = true
			return next, nil
		}
	}
	return "", errors.New("No ip left")
}

func (dhcp *DHCPServer) NextIP() string {
	dhcp.Current++
	dhcp.Current = dhcp.Current % dhcp.Max
	if dhcp.Current == 0 {
		dhcp.Current = 1 // Skip the network address
	}

	curIP := dhcp.Current
	ip := net.IP{0, 0, 0, 0}
	for i := 3; i >= 0; i-- {
		ip[i] = byte(curIP%256) + dhcp.NetworkAddress[i]
		curIP = curIP / 256
	}

	return ip.String()
}

func (dhcp *DHCPServer) Release(ip string) {
	delete(dhcp.Used, ip)
}
