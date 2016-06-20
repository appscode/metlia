package metlia

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"net"

	"github.com/facebookgo/ganglia/gmetric"
	"github.com/rcrowley/go-metrics"
)

type reporter struct {
	// Reporter Configurations.
	config *Config

	// gmetric Client used to communicate with Ganglia.
	client *gmetric.Client

	// Mutex Lock to call the run() only once for multiple Run() called
	// with same registry and configs.
	once   sync.Once
}

func New(conf *Config) (*reporter, error) {
	if conf.Ganglia == nil {
		return nil, errors.New("no ganglia configuration found")
	}

	c, err := conf.Ganglia.newClient()
	if err != nil {
		return nil, err
	}

	if conf.Interval == time.Duration(0) {
		conf.Interval = time.Second * 10
	}

	if conf.PingInterval == time.Duration(0) {
		conf.PingInterval = time.Second * 5
	}

	return &reporter{
		config: conf,
		client: c,
	}, nil
}

// Creates a New gmetric client based on the `gmetric` configs
func (i *Ganglia) newClient() (*gmetric.Client, error) {
	client := &gmetric.Client{
		Addr: []net.Addr{
			&net.UDPAddr{IP: net.ParseIP(i.IP), Port: i.Port},
		},
	}
	if err := client.Open(); err != nil {
		return nil, err
	}
	return client, nil
}

func (r *reporter) Run() {
	r.once.Do(func() {
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					r.handlePanic(rec)
				}
			}()
			r.run()
		}()
	})
}

func (r *reporter) run() {
	intervalTicker := time.Tick(r.config.Interval)
	pingTicker := time.Tick(r.config.PingInterval)

	for {
		select {
		case <-intervalTicker:
			if err := r.send(); err != nil {
				log.Printf("unable to send metrics to ganglia. err=%v", err)
			}
		case <-pingTicker:

		}
	}
}

func (r *reporter) send() error {
	r.config.Registry.Each(func(name string, i interface{}) {
		switch m := i.(type) {
		case metrics.Counter:
			r.reportCounter(name, m)
		case metrics.Histogram:
			r.reportHistogram(name, m)
		case metrics.Gauge:
			r.reportGauge(name, m)
		}
	})
	return nil
}

func (r *reporter) reportCounter(name string, m metrics.Counter) error {
	metric := new(gmetric.Metric)
	metric.TickInterval = 20 * time.Second
	metric.Lifetime = 24 * time.Hour
	metric.Name = fmt.Sprintf("%s.count", name)
	metric.ValueType = gmetric.ValueUint32
	metric.Slope = gmetric.SlopePositive
	return r.sendGangliaMetric(metric, m.Count())
}

func (r *reporter) reportGauge(name string, m metrics.Counter) error {

}

func (r *reporter) getModelMetric(prefix string, name string, valueType string, slope string) *gmetric.Metric {
	metric := new(gmetric.Metric)
	metric.TickInterval = 20 * time.Second
	metric.Lifetime = 24 * time.Hour

	metric.Name = fmt.Sprintf("%s.%s", prefix, name)

	if valueType == "int32" {
		metric.ValueType = gmetric.ValueInt32
	} else if valueType == "float32" {
		metric.ValueType = gmetric.ValueFloat32
	}

	if slope == "positive" {
		metric.Slope = gmetric.SlopePositive
	} else if slope == "both" {
		metric.Slope = gmetric.SlopeBoth
	}

	return metric
}

func (r *reporter) reportHistogram(name string, histogram metrics.Histogram) error {

	metric := r.getModelMetric(name, "max", "int32", "positive")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Max()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "min", "int32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Min()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "stddev", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().StdDev()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "varience", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Variance()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "10thparcentile", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.10)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "25thparcentile", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.25)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "50thparcentile", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.50)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "75thparcentile", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.75)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "99thparcentile", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.99)); err != nil {
		return err
	}
	return nil
}

func (r *reporter) sendGangliaMetric(metric *gmetric.Metric, val interface{}) error {
	if err := r.client.WriteMeta(metric); err != nil {
		return err
	}
	if err := r.client.WriteValue(metric, val); err != nil {
		return err
	}
	return nil
}

func (r *reporter) handlePanic(rec interface{}) {
	logPanic(rec)

	// Additional panic handlers to run
	for _, f := range r.config.PanicHandlers {
		f(r)
	}
}

func logPanic(r interface{}) {
	callers := ""
	for i := 2; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	log.Printf("Recovered from panic: %#v \n%v", r, callers)
}
