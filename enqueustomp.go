package enqueuestomp

import (
	"errors"
	"fmt"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/go-stomp/stomp"
)

func NewEnqueueStomp(c Config) (*EnqueueStomp, error) {
	c.init()

	emq := &EnqueueStomp{
		config: c,
		wp:     workerpool.New(c.MaxWorkers),
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
	opt.init()

	if len(body) == 0 {
		return ErrEmptyPayload
	}

	emq.wp.Submit(func() {
		initTime := time.Now()

		if opt.BeforeSend != nil {
			opt.BeforeSend(queue, initTime)
		}

		destination := fmt.Sprintf("/queue/%s", queue)
		err := emq.conn.Send(destination, opt.ContentType, body, opt.FrameOpts...)

		if errors.Is(err, stomp.ErrAlreadyClosed) || errors.Is(err, stomp.ErrClosedUnexpectedly) {
			emq.conn, _ = emq.newConn()
		}

		if opt.AfterSend != nil {
			opt.AfterSend(queue, initTime, err)
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

	conn, err = stomp.Dial(emq.config.Network, emq.config.Addr, emq.config.ConnOptions...)
	return conn, err
}

func (emq *EnqueueStomp) QueueSize() int {
	return emq.wp.WaitingQueueSize()
}

func (emq *EnqueueStomp) Config() Config {
	return emq.config
}
