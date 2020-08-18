/*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */
package enqueuestomp

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gammazero/workerpool"
	"github.com/go-stomp/stomp"
	uuid "github.com/satori/go.uuid"
)

var (
	ErrEmptyPayload = errors.New("empty payload")
	ErrEmptyQueue   = errors.New("empty queue")
	ErrEmptyTopic   = errors.New("empty topic")
)

type EnqueueStomp struct {
	id           string
	config       Config
	mu           sync.RWMutex
	conn         *stomp.Conn
	wp           *workerpool.WorkerPool
	circuitNames map[string]string
	connected    int32
}

func NewEnqueueStomp(config Config) (*EnqueueStomp, error) {
	config.init()

	emq := &EnqueueStomp{
		id:           makeIdentifier(),
		config:       config,
		wp:           workerpool.New(config.MaxWorkers),
		circuitNames: make(map[string]string),
	}

	if err := emq.newConn(emq.id); err != nil {
		return nil, err
	}

	return emq, nil
}

// SendQueue
// The body array contains the message body,
// and its content should be consistent with the specified content type.
func (emq *EnqueueStomp) SendQueue(queue string, body []byte, so SendOptions) error {
	if queue == "" || strings.TrimSpace(queue) == "" {
		return ErrEmptyQueue
	}
	return emq.send(destinationTypeQueue, queue, body, so)
}

// SendTopic
// The body array contains the message body,
// and its content should be consistent with the specified content type.
func (emq *EnqueueStomp) SendTopic(topic string, body []byte, so SendOptions) error {
	if topic == "" || strings.TrimSpace(topic) == "" {
		return ErrEmptyTopic
	}

	return emq.send(destinationTypeTopic, topic, body, so)
}

func (emq *EnqueueStomp) QueueSize() int {
	return emq.wp.WaitingQueueSize()
}

func (emq *EnqueueStomp) Config() Config {
	return emq.config
}

func (emq *EnqueueStomp) send(destinationType string, destinationName string, body []byte, so SendOptions) error {
	if len(body) == 0 {
		return ErrEmptyPayload
	}
	so.init()
	identifier := makeIdentifier()

	emq.wp.Submit(func() {
		var err error
		startTime := time.Now()
		destination := fmt.Sprintf("/%s/%s", destinationType, destinationName)
		circuitName := emq.makeCircuitName(so.CircuitName)

		if so.BeforeSend != nil {
			so.BeforeSend(identifier, destinationType, destinationName, body, startTime)
		}

	Retry:
		if emq.hasCircuitBreaker(circuitName) {
			err = emq.sendWithCircuitBreaker(circuitName, destination, body, so)
		} else {
			Printf(identifier, "send message")
			err = emq.conn.Send(destination, so.ContentType, body, so.FrameOpts...)
		}

		switch {
		case errors.Is(err, hystrix.ErrCircuitOpen):
			Printf(identifier, "Circuit Open `%s`", err)

		case errors.Is(err, stomp.ErrAlreadyClosed) || errors.Is(err, stomp.ErrClosedUnexpectedly):
			Printf(identifier, "ErrAlreadyClosed `%s`", err)
			atomic.AddInt32(&emq.connected, 0)
			if err = emq.newConn(identifier); err == nil {
				Printf(identifier, "Retry send...")
				goto Retry
			}

		case err != nil:
			logger.Printf("[EnqueueStomp] send error `%s`\n ", err)
		}

		if so.AfterSend != nil {
			so.AfterSend(identifier, destinationType, destinationName, body, startTime, err)
		}
	})

	return nil
}

// NewConn Creates a new conn to broker.
func (emq *EnqueueStomp) newConn(identifier string) (err error) {
	emq.mu.Lock()
	defer emq.mu.Unlock()
	if atomic.LoadInt32(&emq.connected) != 0 {
		return nil
	}

	var conn *stomp.Conn
	for i := 1; i <= emq.config.RetriesConnect; i++ {
		conn, err = stomp.Dial(emq.config.Network, emq.config.Addr, emq.config.ConnOptions...)
		if err == nil && conn != nil {
			Printf(identifier, "Connected :: %s ", emq.config.Addr)
			emq.conn = conn
			atomic.AddInt32(&emq.connected, 1)
			return nil
		}

		timeSleep := emq.config.BackoffConnect(i)
		Printf(identifier,
			"IS OUT OF SERVICE :: %s :: A new conn was tried but failed - sleeping  %s - %d/%d...",
			emq.config.Addr, timeSleep.String(), i, emq.config.RetriesConnect,
		)
		time.Sleep(timeSleep)
	}

	Printf(identifier,
		"IS OUT OF SERVICE :: %s ::",
		emq.config.Addr,
	)

	return err
}

func makeIdentifier() string {
	return uuid.NewV4().String()
}
