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
	c.Assert(config.Options, check.IsNil)
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
	c.Assert(config.Options, check.IsNil)
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
	c.Assert(config.Options, check.IsNil)
	for i := 1; i < 3; i++ {
		c.Assert(config.BackoffConnect(i), check.Equals, enqueuestomp.DefaultInitialBackOff)
	}
}

func (s *EnqueueStompSuite) TestConfigRetriesConnect(c *check.C) {
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
func (s *EnqueueStompSuite) TestConfigWithOptions(c *check.C) {
	enqueueConfig := enqueuestomp.Config{}
	enqueueConfig.AddOptions(
		stomp.ConnOpt.HeartBeat(0*time.Second, 0*time.Second),
	)
	enqueue, err := enqueuestomp.NewEnqueueStomp(enqueueConfig)
	c.Assert(err, check.IsNil)

	config := enqueue.Config()
	c.Assert(config.Options, check.NotNil)
}

func (s *EnqueueStompSuite) TestConfigWithWriteOutputPathInvalid(c *check.C) {
	_, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			WriteOutputPath: "/OutputPathDir/enqueuestomp/enqueuestomp.out",
		},
	)
	c.Assert(err, check.NotNil)
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

	sc := enqueuestomp.SendConfig{}
	err = enqueue.SendQueue(queueName, queueBody, sc)
	c.Assert(err, check.IsNil)
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatQueue(queueName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, "1")
}

func (s *EnqueueStompSuite) TestSendQueueWithBeforeAfterSend(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	sc := enqueuestomp.SendConfig{
		BeforeSend: func(identifier string, destinationType string, destinationName string, body []byte, startTime time.Time) {
			c.Assert(identifier, check.NotNil)
			c.Assert(destinationType, check.Equals, enqueuestomp.DestinationTypeQueue)
			c.Assert(destinationName, check.Equals, queueName)
			c.Assert(string(body), check.Equals, string(queueBody))
			c.Assert(startTime, check.NotNil)
		},
		AfterSend: func(identifier string, destinationType string, destinationName string, body []byte, startTime time.Time, err error) {
			c.Assert(identifier, check.NotNil)
			c.Assert(destinationType, check.Equals, enqueuestomp.DestinationTypeQueue)
			c.Assert(destinationName, check.Equals, queueName)
			c.Assert(string(body), check.Equals, string(queueBody))
			c.Assert(startTime, check.NotNil)
			c.Assert(err, check.IsNil)
		},
	}
	err = enqueue.SendQueue(queueName, queueBody, sc)
	c.Assert(err, check.IsNil)
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatQueue(queueName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, "1")
}

func (s *EnqueueStompSuite) TestSendQueueBodyEmpty(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	sc := enqueuestomp.SendConfig{}
	err = enqueue.SendQueue(queueName, nil, sc)
	c.Assert(err, check.Equals, enqueuestomp.ErrEmptyBody)
}

func (s *EnqueueStompSuite) TestSendTopicBodyEmpty(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	sc := enqueuestomp.SendConfig{}
	err = enqueue.SendTopic(topicName, nil, sc)
	c.Assert(err, check.Equals, enqueuestomp.ErrEmptyBody)
}

func (s *EnqueueStompSuite) TestSendQueueNameEmpty(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	sc := enqueuestomp.SendConfig{}
	err = enqueue.SendQueue("", queueBody, sc)
	c.Assert(err, check.Equals, enqueuestomp.ErrEmptyQueueName)
}

func (s *EnqueueStompSuite) TestSendTopicNameEmpty(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	sc := enqueuestomp.SendConfig{}
	err = enqueue.SendTopic("", topicBody, sc)
	c.Assert(err, check.Equals, enqueuestomp.ErrEmptyTopicName)
}

func (s *EnqueueStompSuite) TestSendQueueWritePersistent(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{},
	)
	c.Assert(err, check.IsNil)

	sc := enqueuestomp.SendConfig{}
	sc.AddOptions(stomp.SendOpt.Header("persistent", "true"))

	err = enqueue.SendQueue(queueName, queueBody, sc)
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

	sc := enqueuestomp.SendConfig{
		CircuitName: "circuit-enqueuestomp",
	}
	err = enqueue.SendQueue(queueName, queueBody, sc)
	c.Assert(err, check.IsNil)
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatQueue(queueName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, "1")
}

func (s *EnqueueStompSuite) TestSendQueueWithWriteDisk(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			WriteOutputPath: queueWriteOutputPath,
		},
	)
	c.Assert(err, check.IsNil)

	total := 100
	for i := 0; i < total; i++ {
		sc := enqueuestomp.SendConfig{}
		err = enqueue.SendQueue(queueName, queueBody, sc)
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
		sc := enqueuestomp.SendConfig{}
		err = enqueue.SendTopic(topicName, topicBody, sc)
		c.Assert(err, check.IsNil)
	}
	s.waitQueueSize(enqueue)

	enqueueCount := s.j.StatTopic(topicName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, strconv.Itoa(total))
}

func (s *EnqueueStompSuite) TestSendTopicWithWriteDisk(c *check.C) {
	enqueue, err := enqueuestomp.NewEnqueueStomp(
		enqueuestomp.Config{
			WriteOutputPath: topicWriteOutputPath,
		},
	)
	c.Assert(err, check.IsNil)

	total := 100
	for i := 0; i < total; i++ {
		sc := enqueuestomp.SendConfig{}
		err = enqueue.SendTopic(topicName, topicBody, sc)
		c.Assert(err, check.IsNil)
	}
	s.waitQueueSize(enqueue)

	lines := s.countFileLines(enqueue.Config().WriteOutputPath)
	c.Assert(lines, check.Equals, total*2)

	enqueueCount := s.j.StatTopic(topicName, "EnqueueCount")
	c.Assert(enqueueCount, check.Equals, strconv.Itoa(total))
}
