package enqueuestomp

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/go-stomp/stomp"
)

func NewEnqueueStomp(config Config) (*EnqueueStomp, error) {
	config.init()

	emq := &EnqueueStomp{
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
func (emq *EnqueueStomp) SendQueue(queue string, body []byte, opt SendOptions) error {
	if queue == "" || strings.TrimSpace(queue) == "" {
		return ErrEmptyQueue
	}

	destination := fmt.Sprintf("/queue/%s", queue)
	return emq.send(destination, body, opt)
}

// The body array contains the message body,
// and its content should be consistent with the specified content type.
func (emq *EnqueueStomp) SendTopic(topic string, body []byte, opt SendOptions) error {
	if topic == "" || strings.TrimSpace(topic) == "" {
		return ErrEmptyTopic
	}

	destination := fmt.Sprintf("/topic/%s", topic)
	return emq.send(destination, body, opt)
}

func (emq *EnqueueStomp) send(destination string, body []byte, opt SendOptions) error {
	if len(body) == 0 {
		return ErrEmptyPayload
	}

	opt.init()
	emq.wp.Submit(func() {
		initTime := time.Now()

		if opt.BeforeSend != nil {
			opt.BeforeSend(destination, initTime)
		}

		err := emq.conn.Send(destination, opt.ContentType, body, opt.FrameOpts...)

		if errors.Is(err, stomp.ErrAlreadyClosed) || errors.Is(err, stomp.ErrClosedUnexpectedly) {
			emq.conn, _ = emq.newConn()
		}

		if opt.AfterSend != nil {
			opt.AfterSend(destination, initTime, err)
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

		timeSleep := emq.config.Backoff(i)
		log.Printf(
			"[EnqueueStomp] IS OUT OF SERVICE :: %s :: A new conn was tried but failed - sleeping  %s ...",
			emq.config.Addr, timeSleep.String(),
		)
		time.Sleep(timeSleep)
	}

	return nil, err
}

func (emq *EnqueueStomp) QueueSize() int {
	return emq.wp.WaitingQueueSize()
}

func (emq *EnqueueStomp) Config() Config {
	return emq.config
}
