# schnapps

## Features

- DNS server (only A records are supported)
- QMP server
- Random MAC address generator
- Uses libvirt events
- Manages a KVM process, its lifecycle and its configuration
- Simple VM scheduler (backend by a queue)

## Roadmap

- Network gateway (via tap/tun on the host)

## Troubleshoot

### DNS response: lame referral

You might have forgot the `.` at the end of your zone name.
