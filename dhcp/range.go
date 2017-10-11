package dhcp

import (
	"net"
)

func next(ip net.IP) net.IP {
	n := len(ip)
	out := make(net.IP, n)
	copy := false
	for n > 0 {
		n--
		if copy {
			out[n] = ip[n]
			continue
		}
		if ip[n] < 255 {
			out[n] = ip[n] + 1
			copy = true
			continue
		}
		out[n] = 0
	}

	return out
}

func generateIpRangeFromNet(ip net.IP, ipnet net.IPNet) []net.IP {
	ipRange := make([]net.IP, 0)

	iIp := ip.Mask(ipnet.Mask)

	for {
		if !ipnet.Contains(iIp) {
			break
		}

		ipRange = append(ipRange, iIp)

		iIp = next(iIp)
	}

	return ipRange
}
