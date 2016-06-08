package metlia

import (
	"time"

	"github.com/rcrowley/go-metrics"
)

type Config struct {
	*Ganglia

	// Registry holds the metricses. That were needed to be
	// reported to Ganglia.
	Registry metrics.Registry

	// Time inerval between two consicutive Ganglia call to
	// store the matrix value to the DB. If not set Default Will
	// be 10secs.
	Interval time.Duration

	// Time interval to make sure the connection between Ganglia
	// and client are alive. if the ping failed with an error client
	// will try to reconnect to Ganglia, if thus failed will throw
	// panics.
	PingInterval time.Duration

	// List of callback functions that will be invoked after every Ganglia
	// call, with the metrics interface that was used to read as param.
	Callbacks []Callback

	// PanicHandlers are the handlers to call whenever a panic occers.
	PanicHandlers []func(interface{})
}

type Ganglia struct {
	IP   string
	Port int
}

// Callbacks invoked after every metric read. the parameter is the metric
// that was read it self.
type Callback func(interface{})
