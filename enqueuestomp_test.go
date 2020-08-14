package enqueuestomp_test

import (
	"runtime"
	"testing"
	"time"

	"gitlab.globoi.com/janilton/enqueuestomp"
	check "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (s *MySuite) TestDefaultConfig(c *check.C) {
	configEnqueue := enqueuestomp.Config{}
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	if err != nil {
		c.Fatal(err)
	}

	config := enqueue.Config()
	c.Assert(config.Addr, check.Equals, "localhost:61613")
	c.Assert(config.Network, check.Equals, "tcp")
	c.Assert(config.MaxWorkers, check.Equals, runtime.NumCPU())
	c.Assert(config.RetriesConnect, check.Equals, 3)
	c.Assert(config.Backoff, check.NotNil)
	for i := 1; i < 3; i++ {
		c.Assert(config.Backoff(i), check.Equals, time.Duration(i*2)*enqueuestomp.DefaultInitialBackOff)
	}
}

func (s *MySuite) TestConfigLinearBackoff(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "localhost:61613",
		Network:        "tcp",
		MaxWorkers:     1,
		RetriesConnect: 1,
		Backoff:        enqueuestomp.LinearBackoff,
	}
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	if err != nil {
		c.Fatal(err)
	}

	config := enqueue.Config()
	c.Assert(config.Addr, check.Equals, configEnqueue.Addr)
	c.Assert(config.Network, check.Equals, configEnqueue.Network)
	c.Assert(config.MaxWorkers, check.Equals, configEnqueue.MaxWorkers)
	c.Assert(config.RetriesConnect, check.Equals, configEnqueue.RetriesConnect)
	c.Assert(config.Backoff, check.NotNil)
	for i := 1; i < 3; i++ {
		c.Assert(config.Backoff(i), check.Equals, time.Duration(i)*enqueuestomp.DefaultInitialBackOff)
	}
}

func (s *MySuite) TestConfigConstantBackOff(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "localhost:61613",
		Network:        "tcp",
		MaxWorkers:     2,
		RetriesConnect: 1,
		Backoff:        enqueuestomp.ConstantBackOff,
	}
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	if err != nil {
		c.Fatal(err)
	}

	config := enqueue.Config()
	c.Assert(config.Addr, check.Equals, configEnqueue.Addr)
	c.Assert(config.Network, check.Equals, configEnqueue.Network)
	c.Assert(config.MaxWorkers, check.Equals, configEnqueue.MaxWorkers)
	c.Assert(config.RetriesConnect, check.Equals, configEnqueue.RetriesConnect)
	c.Assert(config.Backoff, check.NotNil)
	for i := 1; i < 3; i++ {
		c.Assert(config.Backoff(i), check.Equals, enqueuestomp.DefaultInitialBackOff)
	}
}

func (s *MySuite) TestFailtConnect(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "XPTO:61613",
		RetriesConnect: 1,
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}

func (s *MySuite) TestFailtConnect2(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "localhost:123456789",
		RetriesConnect: 1,
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}

func (s *MySuite) TestFailtConnect3(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Network:        "xpto",
		RetriesConnect: 1,
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}
