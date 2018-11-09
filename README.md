# schnapps

> KVM tooling and API

## Examples

- [simple](./examples/simple)

## Running in Docker

- To be able to use KVM, you need to have the proper unix capabilities.
- If you are using the bridge NIC, you need to use the `host` network mode.

## Features

- DNS server (only A records are supported) ([doc](/docs/dns.md))
- QMP server
- Random MAC address generator ([doc](/docs/id.md))
- Uses libvirt events
- Manages a KVM process, its lifecycle and its configuration ([doc](/docs/vm.md))
- Simple VM scheduler with cluster health monitoring ([doc](/docs/scheduler.md))
- Metadata server ([doc](/docs/metadata.md))
- Custom DHCP server (Ipv4 only) ([doc](/docs/dhcp.md))

## Roadmap

- Network gateway (via tap/tun on the host)

## With Linuxkit

- Assign IP using the metadata server: [xtuc/schnapps-ip-metadata](https://github.com/xtuc/Dockerfiles/tree/master/schnapps-ip-metadata)
