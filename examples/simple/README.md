# Simple example

Spawns 4 VMs.

## Environment setup

Install kvm and bridge-utils:

```sh
apt install kvm bridge-utils
```

Setup the bridge:

```sh
sudo brctl addbr brtest
sudo ip addr add 10.1.0.1/24 dev brtest
sudo ip link set dev brtest up
```

Allow qemu to use it:

```sh
sudo mkdir /etc/qemu/
echo "allow brtest" | sudo tee /etc/qemu/bridge.conf
```

## Build VM image

In Linuxkit's folder, run:

```sh
make build
```
