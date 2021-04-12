/*
* enqueuestomp
*
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

	"github.com/gammazero/workerpool"
	"github.com/go-stomp/stomp"
	"go.uber.org/zap"
)

var (
	ErrEmptyBody      = errors.New("empty body")
	ErrEmptyQueueName = errors.New("empty queue name")
	ErrEmptyTopicName = errors.New("empty topic name")

	DefaultExpiresCheck = 1 * time.Minute
	DefaultBodyCheck    = []byte("PING")
)

type EnqueueStomp struct {
	id           string
	config       Config
	mu           sync.RWMutex
	conn         *stomp.Conn
	wp           *workerpool.WorkerPool
	circuitNames map[string]string
	connected    int32
	hasOutput    bool
	output       *zap.Logger
	log          Logger
}

func NewEnqueueStomp(config Config) (*EnqueueStomp, error) {
	config.init()

	emq := &EnqueueStomp{
		id:           config.IdentifierFunc(),
		config:       config,
		wp:           workerpool.New(config.MaxWorkers),
		circuitNames: make(map[string]string),
		log:          config.Logger,
	}

	// create connect
	if err := emq.newConn(emq.id); err != nil {
		return nil, err
	}

	// create output write on disk
	if err := emq.newOutput(); err != nil {
		return nil, err
	}

	return emq, nil
}

// SendQueue
// The body array contains the message body,
// and its content should be consistent with the specified content type.
func (emq *EnqueueStomp) SendQueue(queueName string, body []byte, sc SendConfig, logField LogField) error {
	if queueName == "" || strings.TrimSpace(queueName) == "" {
		return ErrEmptyQueueName
	}
	return emq.send(DestinationTypeQueue, queueName, body, sc, logField)
}

// SendTopic
// The body array contains the message body,
// and its content should be consistent with the specified content type.
func (emq *EnqueueStomp) SendTopic(topicName string, body []byte, sc SendConfig, logField LogField) error {
	if topicName == "" || strings.TrimSpace(topicName) == "" {
		return ErrEmptyTopicName
	}

	return emq.send(DestinationTypeTopic, topicName, body, sc, logField)
}

func (emq *EnqueueStomp) QueueSize() int {
	return emq.wp.WaitingQueueSize()
}

func (emq *EnqueueStomp) Config() Config {
	return emq.config
}

func (emq *EnqueueStomp) CheckQueue(queueName string) error {
	return emq.check(DestinationTypeQueue, queueName)
}

func (emq *EnqueueStomp) CheckTopic(topicName string) error {
	return emq.check(DestinationTypeTopic, topicName)
}

func (emq *EnqueueStomp) Disconnect() error {
	return emq.conn.Disconnect()
}

func (emq *EnqueueStomp) send(destinationType string, destinationName string, body []byte, sc SendConfig, logField LogField) error {
	if len(body) == 0 {
		return ErrEmptyBody
	}
	sc.init()

	identifier := emq.config.IdentifierFunc()
	emq.writeOutput("before", identifier, destinationType, destinationName, body, logField)

	emq.wp.Submit(func() {
		var err error
		startTime := time.Now()
		destination := fmt.Sprintf("/%s/%s", destinationType, destinationName)

		if sc.BeforeSend != nil {
			sc.BeforeSend(identifier, destinationType, destinationName, body, startTime)
		}

	Retry:
		if emq.hasCircuitBreaker(sc) {
			err = emq.sendWithCircuitBreaker(identifier, destination, body, sc)
		} else {
			emq.debugLogger(
				"[enqueuestomp][%s] Send message with destination: `%s` and body: `%s`",
				identifier, destination, body,
			)
			err = emq.conn.Send(destination, sc.ContentType, body, sc.Options...)
		}

		if errors.Is(err, stomp.ErrAlreadyClosed) || errors.Is(err, stomp.ErrClosedUnexpectedly) {
			emq.errorLogger(
				"[enqueuestomp][%s] Connection error `%s`",
				identifier, err,
			)
			atomic.StoreInt32(&emq.connected, 0)
			if err = emq.newConn(identifier); err == nil {
				emq.debugLogger(
					"[enqueuestomp][%s] Retry send...",
					identifier,
				)
				goto Retry
			}
		}

		emq.writeOutput("after", identifier, destinationType, destinationName, body, logField)
		if sc.AfterSend != nil {
			sc.AfterSend(identifier, destinationType, destinationName, body, startTime, err)
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
		conn, err = stomp.Dial(emq.config.Network, emq.config.Addr, emq.config.Options...)
		if err == nil && conn != nil {
			emq.debugLogger(
				"[enqueuestomp][%s] Connected :: %s",
				identifier, emq.config.Addr,
			)
			if atomic.LoadInt32(&emq.connected) == 0 {
				emq.conn = conn
				atomic.StoreInt32(&emq.connected, 1)
			}
			return nil
		}

		timeSleep := emq.config.BackoffConnect(i)
		emq.errorLogger(
			"[enqueuestomp][%s] Connected :: IS OUT OF SERVICE :: %s :: A new conn was tried but failed - sleeping %s - %d/%d",
			identifier,
			emq.config.Addr, timeSleep.String(), i, emq.config.RetriesConnect,
		)
		time.Sleep(timeSleep)
	}

	emq.errorLogger(
		"[enqueuestomp][%s] IS OUT OF SERVICE :: %s",
		identifier, emq.config.Addr,
	)

	return err
}

func (emq *EnqueueStomp) check(destinationType string, destinationName string) error {
	destination := fmt.Sprintf("/%s/%s", destinationType, destinationName)
	unixTimeInMilliSeconds := time.Now().Add(DefaultExpiresCheck).UnixNano() / int64(time.Millisecond)
	contentType := "text/plain"

	return emq.conn.Send(destination,
		contentType,
		DefaultBodyCheck,
		stomp.SendOpt.Header("persistent", "false"),
		stomp.SendOpt.Header("expires", fmt.Sprintf("%d", unixTimeInMilliSeconds)),
	)
}
