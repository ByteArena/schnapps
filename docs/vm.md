# Virtual machine

We use the KVM cli under the hood, make sure you have it installed on your host before.

Tested on QEMU 2.8.1.

## Example usage

```golang
import (
        "github.com/bytearena/schnapps"
)

[…]

config := vmtypes.VMConfig{
    NICs: []interface{}{
        vmtypes.NICBridge{
            Bridge: "HOST_BRIDGE_NAME",
            MAC:    "GUEST_INTERFACE_MAC",
        },
    },
    Id:            id,
    MegMemory:     2048,
    CPUAmount:     1,
    CPUCoreAmount: 1,
    ImageLocation: server.vmRawImageLocation,
    Metadata:      vmtypes.VMMetadata{},
}

arenaVm := vm.NewVM(config)

startErr := arenaVm.Start()

// Error handling
check(startErr)
```

## Network configuration

All the network configuration types are defined in `github.com/bytearena/schnapps/types`.

See the godoc for more information.
