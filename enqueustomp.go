package enqueuestomp

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gammazero/workerpool"
	"github.com/go-stomp/stomp"
)

var (
	ErrEmptyPayload = errors.New("empty payload")
	ErrEmptyQueue   = errors.New("empty queue")
	ErrEmptyTopic   = errors.New("mpty topic")
)

type EnqueueStomp struct {
	id           int
	config       Config
	rw           sync.RWMutex
	conn         *stomp.Conn
	wp           *workerpool.WorkerPool
	circuitNames map[string]string
}

func NewEnqueueStomp(config Config) (*EnqueueStomp, error) {
	config.init()

	rand.Seed(time.Now().UnixNano())
	emq := &EnqueueStomp{
		id:     rand.Int(), // nolint:gosec
		config: config,
		wp:     workerpool.New(config.MaxWorkers),
	}

	var err error
	if emq.conn, err = emq.newConn(); err != nil {
		return nil, err
	}

	return emq, nil
}

// The body array contains the message body,
// and its content should be consistent with the specified content type.
func (emq *EnqueueStomp) SendQueue(queue string, body []byte, so SendOptions) error {
	if queue == "" || strings.TrimSpace(queue) == "" {
		return ErrEmptyQueue
	}

	destination := fmt.Sprintf("/queue/%s", queue)
	return emq.send(destination, body, so)
}

// The body array contains the message body,
// and its content should be consistent with the specified content type.
func (emq *EnqueueStomp) SendTopic(topic string, body []byte, so SendOptions) error {
	if topic == "" || strings.TrimSpace(topic) == "" {
		return ErrEmptyTopic
	}

	destination := fmt.Sprintf("/topic/%s", topic)
	return emq.send(destination, body, so)
}

func (emq *EnqueueStomp) ConfigureCircuitBreaker(
	name string, timeout, maxConcurrentRequests, requestVolumeThreshold, sleepWindow, errorPercentThreshold int,
) {
	emq.rw.Lock()
	defer emq.rw.Unlock()

	circuitName := emq.makeCircuitName(name)
	hystrix.ConfigureCommand(circuitName, hystrix.CommandConfig{
		Timeout:                timeout,
		MaxConcurrentRequests:  maxConcurrentRequests,
		RequestVolumeThreshold: requestVolumeThreshold,
		SleepWindow:            sleepWindow,
		ErrorPercentThreshold:  errorPercentThreshold,
	})
	emq.circuitNames[circuitName] = circuitName
}

func (emq *EnqueueStomp) QueueSize() int {
	return emq.wp.WaitingQueueSize()
}

func (emq *EnqueueStomp) Config() Config {
	return emq.config
}

func (emq *EnqueueStomp) send(destination string, body []byte, so SendOptions) error {
	if len(body) == 0 {
		return ErrEmptyPayload
	}

	so.init()

	emq.wp.Submit(func() {
		initTime := time.Now()

		if so.BeforeSend != nil {
			so.BeforeSend(destination, body, initTime)
		}

		var err error
		circuitName := emq.makeCircuitName(so.CircuitName)
		if emq.hasCircuitBreaker(circuitName) {
			log.Print("[EnqueueStomp] send with CircuitBreaker\n")

			_ = hystrix.Do(circuitName, func() error {
				return emq.conn.Send(destination, so.ContentType, body, so.FrameOpts...)
			}, func(err error) error {
				log.Printf("fallback error %s\n", err)
				return nil
			})

		} else {
			log.Print("[EnqueueStomp] send normal\n")
			err = emq.conn.Send(destination, so.ContentType, body, so.FrameOpts...)
		}

		if errors.Is(err, stomp.ErrAlreadyClosed) || errors.Is(err, stomp.ErrClosedUnexpectedly) {
			emq.conn, _ = emq.newConn()
		}

		if so.AfterSend != nil {
			so.AfterSend(destination, body, initTime, err)
		}
	})

	return nil
}

// NewConn Creates a new conn to broker.
func (emq *EnqueueStomp) newConn() (conn *stomp.Conn, err error) {
	emq.rw.Lock()
	defer emq.rw.Unlock()
	if conn != nil {
		return conn, nil
	}

	for i := 1; i <= emq.config.RetriesConnect; i++ {
		conn, err = stomp.Dial(emq.config.Network, emq.config.Addr, emq.config.ConnOptions...)
		if err == nil && conn != nil {
			log.Printf(
				"[EnqueueStomp] Connected :: %s ",
				emq.config.Addr,
			)
			return conn, nil
		}

		timeSleep := emq.config.BackoffConnect(i)
		log.Printf(
			"[EnqueueStomp] IS OUT OF SERVICE :: %s :: A new conn was tried but failed - sleeping  %s ...",
			emq.config.Addr, timeSleep.String(),
		)
		time.Sleep(timeSleep)
	}

	return nil, err
}

func (emq *EnqueueStomp) makeCircuitName(name string) string {
	return fmt.Sprintf("%s::%d", name, emq.id)
}

func (emq *EnqueueStomp) hasCircuitBreaker(name string) bool {
	_, found := emq.circuitNames[name]
	return found
}

// func (emq *EnqueueStomp) _send(destination string, body []byte, opt SendOptions) error {
// 	// hystrix.CommandConfig

// 	_ = hystrix.Go("tm.Options.HystrixConfig.Name", func() error {

// 		return nil
// 	}, nil)

// 	return nil
// }
