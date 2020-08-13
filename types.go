package enqueuestomp

import (
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/go-stomp/stomp"
	"github.com/go-stomp/stomp/frame"
)

var ErrEmptyPayload = errors.New("empty payload")

type (

	// https://pkg.go.dev/github.com/go-stomp/stomp
	Config struct {
		// Default is tcp.
		Network string

		// host:port address.
		// Default is localhost:61613.
		Addr string

		// stomp.ConnOpt
		ConnOptions []func(*stomp.Conn) error

		// The maxWorkers parameter specifies the maximum number of workers that can
		// execute tasks concurrently.  When there are no incoming tasks, workers are
		// gradually stopped until there are no remaining workers.
		// Default is runtime.NumCPU().
		MaxWorkers int
	}

	// https://pkg.go.dev/github.com/go-stomp/stomp/frame
	SendOptions struct {
		// The content type should be specified, according to the STOMP specification, but if contentType is an empty
		// string, the message will be delivered without a content-type header entry.
		// Default is text/plain.
		ContentType string

		// Any number of options can be specified in opts. See the examples for usage. Options include whether
		// to receive a RECEIPT, should the content-length be suppressed, and sending custom header entries.
		FrameOpts []func(*frame.Frame) error

		BeforeSend func(queue string, init time.Time)

		AfterSend func(queue string, init time.Time, err error)
	}

	EnqueueStomp struct {
		config Config
		rw     sync.RWMutex
		conn   *stomp.Conn
		wp     *workerpool.WorkerPool
	}
)

func (c *Config) init() {
	if c.Network == "" {
		c.Network = "tcp"
	}
	if c.Addr == "" {
		c.Addr = "localhost:61613"
	}

	if c.MaxWorkers < 1 {
		c.MaxWorkers = runtime.NumCPU()
	}
}

func (so *SendOptions) init() {
	if so.ContentType == "" {
		so.ContentType = "text/plain"
	}
}
