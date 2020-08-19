/*
* enqueuestomp
*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */

package enqueuestomp

import (
	"runtime"

	"github.com/go-stomp/stomp"
)

const (
	defaultRetriesConnect    = 3
	defaultMaxRetriesConnect = 5
)

type Config struct {
	// Default is tcp
	Network string

	// host:port address.
	// Default is localhost:61613
	Addr string

	// stomp.ConnOpt
	// https://pkg.go.dev/github.com/go-stomp/stomp
	Options []func(*stomp.Conn) error

	// The maxWorkers parameter specifies the maximum number of workers that can
	// execute tasks concurrently.  When there are no incoming tasks, workers are
	// gradually stopped until there are no remaining workers.
	// Default is runtime.NumCPU()
	MaxWorkers int

	// Default is 3, Max 5
	RetriesConnect int

	// Default is ExponentialBackoff
	BackoffConnect BackoffStrategy

	// Enabled output file in disk
	// Default is false
	WriteOnDisk bool

	// WriteOutputPath is file path to write logging output to.
	WriteOutputPath string
}

func (c *Config) AddOpts(opts ...func(*stomp.Conn) error) {
	c.Options = opts
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
	} else if c.RetriesConnect > defaultMaxRetriesConnect {
		c.RetriesConnect = defaultMaxRetriesConnect
	}

	if c.BackoffConnect == nil {
		c.BackoffConnect = ExponentialBackoff
	}
}
