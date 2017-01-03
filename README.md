[![CLA assistant](https://cla-assistant.io/readme/badge/appscode/metlia)](https://cla-assistant.io/appscode/metlia)

[Website](https://appscode.com) • [Slack](https://appscode.slack.com) • [Forum](https://discuss.appscode.com) • [Twitter](https://twitter.com/AppsCodeHQ)

Metlia
======

This is a reporter for the [go-metrics](https://github.com/rcrowley/go-metrics) library which will post the metrics to ganglia.

Note
----

[ganglia](https://github.com/facebookgo/ganglia) library is used to send metrics to ganglia.


**Get this package using**
```bash
$ go get github.com/appscode/metlia
```

Usage
-----

```go
import (
	"net"
	"time"

	"github.com/appscode/metlia"
	"github.com/rcrowley/go-metrics"
)

reg := metrics.NewRegistry()
addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8649")

go metlia.Ganglia(reg, 30 * time.Second, addr)
```

License
-------
`metlia` is licensed under the Apache 2.0 license. See the LICENSE file for details.
