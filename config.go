package enqueuestomp

import (
	"runtime"

	"github.com/go-stomp/stomp"
)

const (
	defaultRetriesConnect = 3
)

type Config struct {
	// Default is tcp
	Network string

	// host:port address.
	// Default is localhost:61613
	Addr string

	// stomp.ConnOpt
	ConnOptions []func(*stomp.Conn) error

	// The maxWorkers parameter specifies the maximum number of workers that can
	// execute tasks concurrently.  When there are no incoming tasks, workers are
	// gradually stopped until there are no remaining workers.
	// https://pkg.go.dev/github.com/go-stomp/stomp
	// Default is runtime.NumCPU()
	MaxWorkers int

	// Default is 3
	RetriesConnect int

	// Default is ExponentialBackoff
	BackoffConnect BackoffStrategy
}

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

	if c.RetriesConnect < 1 {
		c.RetriesConnect = defaultRetriesConnect
	}

	if c.BackoffConnect == nil {
		c.BackoffConnect = ExponentialBackoff
	}
}
