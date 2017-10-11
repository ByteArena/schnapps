# Metadata server

You can store metadata into a VM. It's a `map[string]string` where you are free to store whatever information you need. `schnapps` won't use it.

## Example usage

### Server usage

```golang
import (
        "github.com/bytearena/schnapps"
        vmmeta "github.com/bytearena/schnapps/metadata"
)

retrieveVMFn := func(id string) *vm.VM {
	vm := FindVMByMAC(yourState, id)

	return vm
}

metadataServer := vmmeta.NewServer("127.0.0.1:8080", retrieveVMFn)

err := metadataServer.Start()

// Error handling
check(err)
```

The retrieve function will be called for each metadata request from the guests.

### Defining metadata

```golang
import (
        "github.com/bytearena/schnapps"
        vmtypes "github.com/bytearena/schnapps/types"
)

[…]

// Define metadata
metadata := vmtypes.VMMetadata{}

config := vmtypes.VMConfig{
    […]
    Metadata: metadata,
}

myVm := vm.NewVM(config)

// Reading metadata
log.Println(vm.Config.Metadata["something"])
```
