# Domain Name System

The DNS server is optionnal.

Standart ports are `53` and `5353`.

## Example usage

```golang
import (
        vmdns "github.com/bytearena/schnapps/dns"
)

[…]

zone := "sven.com."

records := map[string]string{
    "foo.sven.com.": "1.2.3.4",
    "bar.sven.com.": "4.3.2.1",
}

DNSServer := vmdns.MakeServer("127.0.0.1:53", zone, records)

err := DNSServer.Start()

// Error handling
check(err)
```

## Troubleshoot

### DNS response: lame referral

You might have forgot the `.` at the end of your zone name.
