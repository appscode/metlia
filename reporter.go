package metlia

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/facebookgo/ganglia/gmetric"
	"github.com/rcrowley/go-metrics"
)

type Reporter struct {
	// Address of Ganglia gmond process
	Addr *net.UDPAddr

	// Registry holds the matrices. That were needed to be
	// reported to Ganglia.
	Registry metrics.Registry

	// Time inerval between two consicutive Ganglia call to
	// store the matrix value to the DB. If not set Default Will
	// be 10secs.
	FlushInterval time.Duration

	// gmetric Client used to communicate with Ganglia.
	client *gmetric.Client
}

func Ganglia(r metrics.Registry, d time.Duration, addr *net.UDPAddr) {
	defer func() {
		if rec := recover(); rec != nil {
			handlePanic(rec)
		}
	}()

	for range time.Tick(d) {
		g := &Reporter{
			Addr:          addr,
			Registry:      r,
			FlushInterval: d,
		}
		if err := g.Send(); err != nil {
			log.Println(err)
		}
	}
}

func (r *Reporter) Send() error {
	r.client = &gmetric.Client{
		Addr: []net.Addr{r.Addr},
	}
	if err := r.client.Open(); err != nil {
		return err
	}
	defer r.client.Close()

	r.Registry.Each(func(name string, i interface{}) {
		switch m := i.(type) {
		case metrics.Counter:
			r.reportCounter(name, m.Snapshot())
		case metrics.Histogram:
			r.reportHistogram(name, m.Snapshot())
		case metrics.Gauge:
			r.reportGauge(name, m.Snapshot())
		case metrics.Meter:
			r.reportMeter(name, m.Snapshot())
		case metrics.Timer:
			r.reportTimer(name, m.Snapshot())
		}
	})
	return nil
}

func (r *Reporter) getModelMetric(prefix string, name string, valueType string, slope string) *gmetric.Metric {
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
	} else if slope == "negative" {
		metric.Slope = gmetric.SlopeNegative
	} else if slope == "both" {
		metric.Slope = gmetric.SlopeBoth
	}
	return metric
}

func (r *Reporter) reportCounter(name string, m metrics.Counter) error {
	metric := r.getModelMetric(name, "count", "int32", "positive")
	return r.sendGangliaMetric(metric, m.Count())
}

func (r *Reporter) reportGauge(name string, g metrics.Gauge) error {
	metric := r.getModelMetric(name, "value", "float32", "both")
	return r.sendGangliaMetric(metric, g.Value())
}

func (r *Reporter) reportMeter(name string, m metrics.Meter) error {
	metric := r.getModelMetric(name, "count", "int32", "positive")
	if err := r.sendGangliaMetric(metric, m.Count()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "mean_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, m.RateMean()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "m1_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, m.Rate1()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "m5_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, m.Rate5()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "m15_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, m.Rate15()); err != nil {
		return err
	}
	return nil
}

func (r *Reporter) reportTimer(name string, timer metrics.Timer) error {
	metric := r.getModelMetric(name, "count", "int32", "positive")
	if err := r.sendGangliaMetric(metric, timer.Count()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "max", "int32", "positive")
	if err := r.sendGangliaMetric(metric, timer.Max()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "min", "int32", "both")
	if err := r.sendGangliaMetric(metric, timer.Min()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "mean", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Mean()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "mean_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.RateMean()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "m1_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Rate1()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "m5_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Rate5()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "m15_rate", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Rate15()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "variance", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Variance()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "stddev", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.StdDev()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "sum", "int32", "positive")
	if err := r.sendGangliaMetric(metric, timer.Sum()); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "p50", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Percentile(0.50)); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "p75", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Percentile(0.75)); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "p95", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Percentile(0.95)); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "p98", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Percentile(0.98)); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "p99", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Percentile(0.99)); err != nil {
		return err
	}
	metric = r.getModelMetric(name, "p999", "float32", "both")
	if err := r.sendGangliaMetric(metric, timer.Percentile(0.999)); err != nil {
		return err
	}
	return nil
}

func (r *Reporter) reportHistogram(name string, histogram metrics.Histogram) error {
	metric := r.getModelMetric(name, "max", "int32", "positive")
	if err := r.sendGangliaMetric(metric, histogram.Max()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "min", "int32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Min()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "stddev", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.StdDev()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "varience", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Variance()); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "p50", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Percentile(0.50)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "p75", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Percentile(0.75)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "p95", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Percentile(0.95)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "p98", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Percentile(0.98)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "p99", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Percentile(0.99)); err != nil {
		return err
	}

	metric = r.getModelMetric(name, "p999", "float32", "both")
	if err := r.sendGangliaMetric(metric, histogram.Percentile(0.999)); err != nil {
		return err
	}

	return nil
}

func (r *Reporter) sendGangliaMetric(metric *gmetric.Metric, val interface{}) error {
	if err := r.client.WriteMeta(metric); err != nil {
		return err
	}
	if err := r.client.WriteValue(metric, val); err != nil {
		return err
	}
	return nil
}

func handlePanic(rec interface{}) {
	callers := ""
	for i := 2; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	log.Printf("Recovered from panic: %#v \n%v", rec, callers)
}
