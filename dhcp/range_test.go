package dhcp

import (
	"net"
	"testing"
)

func TestGenerateIpRangeFromNet(t *testing.T) {
	startIp := net.IP{192, 168, 1, 0}
	mask := net.IPMask{255, 255, 255, 254}

	netIp := net.IPNet{
		IP:   startIp,
		Mask: mask,
	}

	ipRange := generateIpRangeFromNet(startIp, netIp)

	if ipRange[0].String() != "192.168.1.0" {
		panic("First ip is not correct")
	}

	if ipRange[1].String() != "192.168.1.1" {
		panic("Last ip is not correct")
	}
}
