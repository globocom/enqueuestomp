package enqueuestomp_test

import (
	"runtime"
	"testing"

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
}

func (s *MySuite) TestConfig(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:       "localhost:61613",
		Network:    "tcp",
		MaxWorkers: 1,
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
}

func (s *MySuite) TestFailtConnect(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr: "XPTO:61613",
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}

func (s *MySuite) TestFailtConnect2(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr: "localhost:123456789",
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}

func (s *MySuite) TestFailtConnect3(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Network: "xpto",
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}
