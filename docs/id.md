# Identification

### Random MAC address

We use local assigned MAC address.
The following pattern is used to generate the address: 00:f0:xx:xx:xx:xx.

#### Example usage

```golang
import (
        vmid "github.com/bytearena/schnapps/id"
)

[…]

mac := vmid.GenerateRandomMAC()
```
