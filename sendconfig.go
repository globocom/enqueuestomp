/*
* enqueuestomp
*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */

package enqueuestomp

import (
	"time"

	"github.com/go-stomp/stomp/frame"
)

const (
	DestinationTypeQueue = "queue"
	DestinationTypeTopic = "topic"
)

type SendConfig struct {
	// The content type should be specified, according to the STOMP specification, but if contentType is an empty
	// string, the message will be delivered without a content-type header entry.
	// Default is text/plain.
	ContentType string

	// Any number of options can be specified in opts. See the examples for usage. Options include whether
	// to receive a RECEIPT, should the content-length be suppressed, and sending custom header entries.
	// https://pkg.go.dev/github.com/go-stomp/stomp/frame
	Options []func(*frame.Frame) error

	BeforeSend func(identifier string, destinationType string, destinationName string, body []byte, startTime time.Time)

	AfterSend func(identifier string, destinationType string, destinationName string, body []byte, startTime time.Time, err error)

	// the name of the CircuitBreaker.
	// Default is empty
	CircuitName string

	logField LogField
}

func (sc *SendConfig) SetOptions(opts ...func(*frame.Frame) error) {
	sc.Options = opts
}

func (sc *SendConfig) AddOption(opt func(*frame.Frame) error) {
	if len(sc.Options) == 0 {
		sc.Options = make([]func(*frame.Frame) error, 0)
	}
	sc.Options = append(sc.Options, opt)
}

func (sc *SendConfig) AddLogField(key, value string) {
	if sc.logField == nil {
		sc.logField = newLogField()
	}
	sc.logField.setNewField(key, value)
}

func (sc *SendConfig) init() {
	if sc.ContentType == "" {
		sc.ContentType = "text/plain"
	}
}
