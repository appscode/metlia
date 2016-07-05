[![CLA assistant](https://cla-assistant.io/readme/badge/appscode/metlia)](https://cla-assistant.io/appscode/metlia)

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
  "github.com/appscode/metlia"
  "github.com/rcrowley/go-metrics"
)

config := &metlia.Config{
  Registry: metrics.NewRegistry(),  // metrics registry
  Interval: time.Second * 60,       // interval
  Ganglia: &metlia.Ganglia{
    IP:   "url",                    // Ganglia url
    Port: 8649,                     // Ganglia port
    },
}

reporter, err := metlia.New(config)
if err == nil {
  reporter.Run()
}
```

License
-------
`metlia` is licensed under the Apache 2.0 license. See the LICENSE file for details.
