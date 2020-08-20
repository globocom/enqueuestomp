# enqueuestomp

[![CircleCI](https://circleci.com/gh/globocom/enqueuestomp.svg?style=shield)](https://circleci.com/gh/globocom/enqueuestomp)
[![Release](https://img.shields.io/github/release/globocom/enqueuestomp.svg)](https://github.com/globocom/enqueuestomp/releases)
[![GoDoc]( https://godoc.org/github.com/globocom/enqueuestomp?status.svg)](https://pkg.go.dev/github.com/globocom/enqueuestomp)

## Use

```go
package main

import (
    "log"

    "github.com/globocom/enqueuestomp"
    "github.com/go-stomp/stomp"
)

func main() {
    enqueueConfig := enqueuestomp.Config{
        Addr:  "localhost:61613",
    }
    enqueueConfig.AddOptions(
        stomp.ConnOpt.HeartBeat(0*time.Second, 0*time.Second),
    )

    enqueue, err := enqueuestomp.NewEnqueueStomp(
        enqueueConfig,
    )
    if err != nil {
        log.Fatalf("error %s", err)
    }

    sc := enqueuestomp.SendConfig{}
    sc.AddOptions(
        stomp.SendOpt.Header("persistent", "true"),
    )

    name := "queueName"
    body := []byte("queueBody")
    err := enqueue.SendQueue(name, body, sc)
    if err != nil {
        fmt.Printf("error %s", err)
    }
}
```

### Enqueue config

```go
    type Config struct {
        // Default is tcp
        Network string

        // host:port address
        // Default is localhost:61613
        Addr string

        // https://pkg.go.dev/github.com/go-stomp/stomp
        Options []func(*stomp.Conn) error

        // The maxWorkers parameter specifies the maximum number of workers that can
        // execute tasks concurrently.  When there are no incoming tasks, workers are
        // gradually stopped until there are no remaining workers.
        // Default is runtime.NumCPU()
        MaxWorkers int

        // Default is 3, Max is 5
        RetriesConnect int

        // Used to determine how long a retry request should wait until attempted.
        // Default is ExponentialBackoff
        BackoffConnect BackoffStrategy

        // File path to write logging output to
        WriteOutputPath string

        // Logger that will be used
        // Default is nothing
        Logger Logger
    }
```

### Send config

```go
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
    }
```

### Documentation

[![GoDoc]( https://godoc.org/github.com/globocom/enqueuestomp?status.svg)](https://pkg.go.dev/github.com/globocom/enqueuestomp)

Full documentation for the package can be viewed online using the GoDoc site here:
https://pkg.go.dev/github.com/globocom/enqueuestomp
