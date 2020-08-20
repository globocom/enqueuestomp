/*
* enqueuestomp
*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */

package enqueuestomp_test

import (
	"runtime"
	"strconv"
	"time"

	"github.com/globocom/enqueuestomp"
	"github.com/go-stomp/stomp"
	check "gopkg.in/check.v1"
)

var (
	// Queue.
	queueName            = "testQueue"
	queueBody            = []byte("bodyQueue")
	queueWriteOutputPath = "enqueuestompQueue.log"

	// Topic.
	topicName            = "testTopic"
	topicBody            = []byte("bodyTopic")
	topicWriteOutputPath = "enqueuestompTopic.log"
)

func (s *EnqueueStompSuite) TestDefaultConfig(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
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

func (s *EnqueueStompSuite) TestConfigetriesConnect(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			RetriesConnect: 1,
		},
	)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.RetriesConnect, check.Equals, 1)
}

func (s *EnqueueStompSuite) TestConfigMinRetriesConnect(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			RetriesConnect: -1,
		},
	)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.RetriesConnect, check.Equals, 3)
}

func (s *EnqueueStompSuite) TestConfigMinRetriesConnect2(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			RetriesConnect: 0,
		},
	)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.RetriesConnect, check.Equals, 3)
}

func (s *EnqueueStompSuite) TestConfigMaxRetriesConnect(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			RetriesConnect: 10,
		},
	)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.RetriesConnect, check.Equals, 5)
}

func (s *EnqueueStompSuite) TestFailtConnectAddr(c *check.C) {
	_, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			Addr:           "notfound:1234",
			RetriesConnect: 1,
		},
	)
	c.Assert(err, check.NotNil)
}

func (s *EnqueueStompSuite) TestFailtConnectNetwork(c *check.C) {
	_, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			Network:        "xpto",
			RetriesConnect: 1,
		},
	)
	c.Assert(err, check.NotNil)
}

func (s *EnqueueStompSuite) TestSendQueue(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	so := enqueuestomp.SendOptions{}
	err = enqueue.SendQueue(queueName, queueBody, so)
	c.Assert(err, check.IsNil)
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatQueue(queueName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, "1")
}

func (s *EnqueueStompSuite) TestSendQueueWritePersistent(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	so := enqueuestomp.SendOptions{}
	so.AddOptions(stomp.SendOpt.Header("persistent", "true"))

	err = enqueue.SendQueue(queueName, queueBody, so)
	c.Assert(err, check.IsNil)
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatQueue(queueName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, "1")
}

func (s *EnqueueStompSuite) TestSendQueueWithCircuitBreaker(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	enqueue.ConfigureCircuitBreaker(
		"circuit-enqueuestomp",
		enqueuestomp.CircuitBreakerConfig{},
	)

	so := enqueuestomp.SendOptions{
		CircuitName: "circuit-enqueuestomp",
	}
	err = enqueue.SendQueue(queueName, queueBody, so)
	c.Assert(err, check.IsNil)
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatQueue(queueName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, "1")
}

func (s *EnqueueStompSuite) TestSendQueueWithWriteDisk(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			WriteOnDisk:     true,
			WriteOutputPath: queueWriteOutputPath,
		},
	)
	c.Assert(err, check.IsNil)

	total := 100
	for i := 0; i < total; i++ {
		so := enqueuestomp.SendOptions{}
		err = enqueue.SendQueue(queueName, queueBody, so)
		c.Assert(err, check.IsNil)
	}
	s.waitQueueSize(enqueue)

	lines := s.countFileLines(enqueue.Config().WriteOutputPath)
	c.Assert(lines, check.Equals, total*2)

	enqueueCount := s.j.StatQueue(queueName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, strconv.Itoa(total))
}

func (s *EnqueueStompSuite) TestSendTopic(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	total := 100
	for i := 0; i < total; i++ {
		so := enqueuestomp.SendOptions{}
		err = enqueue.SendTopic(topicName, topicBody, so)
		c.Assert(err, check.IsNil)
	}
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatTopic(topicName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, strconv.Itoa(total))
}

func (s *EnqueueStompSuite) TestSendTopicWithWriteDisk(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			WriteOnDisk:     true,
			WriteOutputPath: topicWriteOutputPath,
		},
	)
	c.Assert(err, check.IsNil)

	total := 100
	for i := 0; i < total; i++ {
		so := enqueuestomp.SendOptions{}
		err = enqueue.SendTopic(topicName, topicBody, so)
		c.Assert(err, check.IsNil)
	}
	s.waitQueueSize(enqueue)

	lines := s.countFileLines(enqueue.Config().WriteOutputPath)
	c.Assert(lines, check.Equals, total*2)

	enqueueCount := s.j.StatTopic(topicName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, strconv.Itoa(total))
}
