package dhcp

import (
	"errors"
	"net"
)

const (
	IPAvailable = true
)

type ipRangeBackend map[*net.IP]bool

type DHCPServer struct {
	ips ipRangeBackend
}

func NewDHCPServer(cidr string) (*DHCPServer, error) {
	ipv4Addr, ipv4Net, err := net.ParseCIDR(cidr)

	if err != nil {
		return nil, err
	}

	ips := generateIpRangeFromNet(ipv4Addr, *ipv4Net)
	backend := make(ipRangeBackend, 0)

	for _, ip := range ips {
		backend[&ip] = IPAvailable
	}

	return &DHCPServer{
		ips: backend,
	}, nil
}

func (dhcp *DHCPServer) Pop() (string, error) {
	if len(dhcp.ips) == 0 {
		return "", errors.New("Cannot pop from pool: no more IP available")
	}

	var ip string

	// Take the first one (in guarantee random order)
	for x, _ := range dhcp.ips {
		ip = x.String()
		delete(dhcp.ips, x)

		break
	}

	return ip, nil
}

func (dhcp *DHCPServer) Release(ip string) {
	parsedIp := net.ParseIP(ip)

	dhcp.ips[&parsedIp] = IPAvailable
}
