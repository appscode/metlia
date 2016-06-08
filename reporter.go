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
			r.reportHistogram(m, name)
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

func (r *reporter) reportHistogram(histogram metrics.Histogram, name string) error {
	metric := new(gmetric.Metric)
	metric.TickInterval = 20 * time.Second
	metric.Lifetime = 24 * time.Hour

	metric.Name = fmt.Sprintf("%s.max", name)
	metric.ValueType = gmetric.ValueUint32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Max()); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.min", name)
	metric.ValueType = gmetric.ValueUint32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Min()); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.stddev", name)
	metric.ValueType = gmetric.ValueFloat32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().StdDev()); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.variance", name)
	metric.ValueType = gmetric.ValueFloat32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Variance()); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.10thparcentile", name)
	metric.ValueType = gmetric.ValueFloat32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.10)); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.25thparcentile", name)
	metric.ValueType = gmetric.ValueFloat32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.25)); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.50thparcentile", name)
	metric.ValueType = gmetric.ValueFloat32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.50)); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.75thparcentile", name)
	metric.ValueType = gmetric.ValueFloat32
	metric.Slope = gmetric.SlopeBoth
	if err := r.sendGangliaMetric(metric, histogram.Snapshot().Percentile(0.75)); err != nil {
		return err
	}

	metric.Name = fmt.Sprintf("%s.99thparcentile", name)
	metric.ValueType = gmetric.ValueFloat32
	metric.Slope = gmetric.SlopeBoth
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
