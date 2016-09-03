package metlia_test

import (
	"net"
	"testing"
	"time"

	"github.com/appscode/metlia"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	reg := metrics.NewRegistry()
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8649")
	assert.Nil(t, err)
	g := &metlia.Reporter{
		Addr:          addr,
		Registry:      reg,
		FlushInterval: 1 * time.Second,
	}

	for i := 1; i <= 5; i++ {
		con1 := reg.GetOrRegister("conn1", metrics.NewCounter()).(metrics.Counter)
		con2 := reg.GetOrRegister("conn2", metrics.NewCounter()).(metrics.Counter)

		con1.Inc(1)
		con2.Inc(5)

		g.Send()
		time.Sleep(time.Second * 1)
	}
}
func TestGauge(t *testing.T) {
	reg := metrics.NewRegistry()
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8649")
	assert.Nil(t, err)
	g := &metlia.Reporter{
		Addr:          addr,
		Registry:      reg,
		FlushInterval: 1 * time.Second,
	}

	for i := 1; i <= 20; i++ {
		con1 := reg.GetOrRegister("conn3", metrics.NewGauge()).(metrics.Gauge)
		con1.Update(2)
		g.Send()
		time.Sleep(time.Second * 1)
	}
}

func TestMeter(t *testing.T) {
	reg := metrics.NewRegistry()
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8649")
	assert.Nil(t, err)
	g := &metlia.Reporter{
		Addr:          addr,
		Registry:      reg,
		FlushInterval: 1 * time.Second,
	}

	for i := 1; i <= 5; i++ {
		con5 := reg.GetOrRegister("conn5", metrics.NewMeter()).(metrics.Meter)
		con5.Mark(int64(i))
		g.Send()
		time.Sleep(time.Second * 1)
	}
}

func TestTimer(t *testing.T) {
	reg := metrics.NewRegistry()
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8649")
	assert.Nil(t, err)
	g := &metlia.Reporter{
		Addr:          addr,
		Registry:      reg,
		FlushInterval: 1 * time.Second,
	}

	for i := 1; i <= 5; i++ {
		con6 := reg.GetOrRegister("conn6", metrics.NewTimer()).(metrics.Timer)
		con6.Update(time.Duration(100 * i))
		g.Send()
		time.Sleep(time.Second * 1)
	}
}

func TestHistogram(t *testing.T) {
	reg := metrics.NewRegistry()
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8649")
	assert.Nil(t, err)
	g := &metlia.Reporter{
		Addr:          addr,
		Registry:      reg,
		FlushInterval: 1 * time.Second,
	}

	for i := 1; i <= 5; i++ {
		con7 := reg.GetOrRegister("conn7", metrics.NewHistogram(metrics.NewUniformSample(100))).(metrics.Histogram)
		con7.Update(int64(i * 2))
		g.Send()
		time.Sleep(time.Second * 1)
	}
}
