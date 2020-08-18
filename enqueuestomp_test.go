package enqueuestomp_test

import (
	"runtime"
	"time"

	"github.com/globocom/enqueuestomp"
	check "gopkg.in/check.v1"
)

func (s *EnqueueStompSuite) TestDefaultConfig(c *check.C) {
	configEnqueue := enqueuestomp.Config{}
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.Addr, check.Equals, "localhost:61613")
	c.Assert(config.Network, check.Equals, "tcp")
	c.Assert(config.MaxWorkers, check.Equals, runtime.NumCPU())
	c.Assert(config.RetriesConnect, check.Equals, 3)
	c.Assert(config.BackoffConnect, check.NotNil)
	for i := 1; i < 3; i++ {
		c.Assert(config.BackoffConnect(i), check.Equals, time.Duration(i*2)*enqueuestomp.DefaultInitialBackOff)
	}
}

func (s *EnqueueStompSuite) TestConfigLinearBackoff(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "localhost:61613",
		Network:        "tcp",
		MaxWorkers:     1,
		RetriesConnect: 1,
		BackoffConnect: enqueuestomp.LinearBackoff,
	}
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.Addr, check.Equals, configEnqueue.Addr)
	c.Assert(config.Network, check.Equals, configEnqueue.Network)
	c.Assert(config.MaxWorkers, check.Equals, configEnqueue.MaxWorkers)
	c.Assert(config.RetriesConnect, check.Equals, configEnqueue.RetriesConnect)
	c.Assert(config.BackoffConnect, check.NotNil)
	for i := 1; i < 3; i++ {
		c.Assert(config.BackoffConnect(i), check.Equals, time.Duration(i)*enqueuestomp.DefaultInitialBackOff)
	}
}

func (s *EnqueueStompSuite) TestConfigConstantBackOff(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "localhost:61613",
		Network:        "tcp",
		MaxWorkers:     2,
		RetriesConnect: 1,
		BackoffConnect: enqueuestomp.ConstantBackOff,
	}
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.Addr, check.Equals, configEnqueue.Addr)
	c.Assert(config.Network, check.Equals, configEnqueue.Network)
	c.Assert(config.MaxWorkers, check.Equals, configEnqueue.MaxWorkers)
	c.Assert(config.RetriesConnect, check.Equals, configEnqueue.RetriesConnect)
	c.Assert(config.BackoffConnect, check.NotNil)
	for i := 1; i < 3; i++ {
		c.Assert(config.BackoffConnect(i), check.Equals, enqueuestomp.DefaultInitialBackOff)
	}
}

func (s *EnqueueStompSuite) TestFailtConnect(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "XPTO:61613",
		RetriesConnect: 1,
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}

func (s *EnqueueStompSuite) TestFailtConnect2(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Addr:           "localhost:123456789",
		RetriesConnect: 1,
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}

func (s *EnqueueStompSuite) TestFailtConnect3(c *check.C) {
	configEnqueue := enqueuestomp.Config{
		Network:        "xpto",
		RetriesConnect: 1,
	}
	_, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.NotNil)
}

func (s *EnqueueStompSuite) TestSendQueue(c *check.C) {
	configEnqueue := enqueuestomp.NewConfig("localhost:61613", 1, 1)
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.IsNil)

	queueName := "testeStomp"
	body := []byte("bodyStomp")
	so := enqueuestomp.SendOptions{}
	err = enqueue.SendQueue(queueName, body, so)

	c.Assert(err, check.IsNil)
}

func (s *EnqueueStompSuite) TestSendQueueWithCircuitBreaker(c *check.C) {
	configEnqueue := enqueuestomp.NewConfig("localhost:61613", 1, 1)
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		configEnqueue,
	)
	c.Assert(err, check.IsNil)

	enqueue.ConfigureCircuitBreaker(
		"circuit-enqueuestomp",
		enqueuestomp.CircuitBreakerConfig{},
	)

	queueName := "testeStomp"
	body := []byte("bodyStomp")
	so := enqueuestomp.SendOptions{
		CircuitName: "circuit-enqueuestomp",
	}
	err = enqueue.SendQueue(queueName, body, so)

}
