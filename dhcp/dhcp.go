package dhcp

import (
	"errors"
	"net"
)

const (
	IPAvailable = true
)

type key struct {
	ip net.IP
}

type ipRangeBackend map[*key]bool

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
		key := key{ip}
		backend[&key] = IPAvailable
	}

	return &DHCPServer{
		ips: backend,
	}, nil
}

func (dhcp *DHCPServer) Pop() (string, error) {
	if len(dhcp.ips) == 0 {
		return "", errors.New("Cannot pop from pool: no more IP available")
	}

	// Take the first one (in guarantee random order)
	for x, _ := range dhcp.ips {
		delete(dhcp.ips, x)

		return x.ip.String(), nil
	}

	return "", nil
}

func (dhcp *DHCPServer) Release(ip string) {
	parsedIp := net.ParseIP(ip)
	key := key{parsedIp}

	dhcp.ips[&key] = IPAvailable
}
