# Dynamic Host Configuration Protocol

The DHCP server is optionnal. It doesn't expose a real DHCP server at the moment, we use it to attribute IP and communicate them via the metadata server.

Only Ipv4 is supported.

## Example usage

```golang
import (
        vmdhcp "github.com/bytearena/schnapps/dhcp"
)

[…]

cidr := "192.168.0.1/24"
server, err := vmdhcp.NewDHCPServer(cidr)

ip, err := server.Pop()

// Error handling, because pool can be empty
check(err)

// Release the ip
server.Release(ip)
```
