package metlia_test

import (
	"testing"
	"time"

	"github.com/appscode/metlia"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

func TestReporter(t *testing.T) {
	reg := metrics.NewRegistry()
	config := &metlia.Config{
		Ganglia: &metlia.Ganglia{
			IP:   "",
			Port: 8649,
		},
		Registry: reg,
		Interval: time.Second * 5,
	}

	reporter, err := metlia.New(config)
	assert.Nil(t, err)

	reporter.Run()

	for i := 1; i <= 20; i++ {
		con1 := reg.GetOrRegister("conn1", metrics.NewCounter()).(metrics.Counter)
		con2 := reg.GetOrRegister("conn2", metrics.NewCounter()).(metrics.Counter)

		con1.Inc(1)
		con2.Inc(5)

		time.Sleep(time.Second * 10)
	}
}
