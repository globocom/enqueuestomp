package enqueuestomp

import (
	"fmt"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	defaultTimeout                = 10000
	defaultMaxConcurrentRequests  = 10000
	defaultRequestVolumeThreshold = 100
	defaultSleepWindow            = 500
	defaultErrorPercentThreshold  = 5
)

type CircuitBreakerConfig struct {
	// how long to wait for command to complete, in milliseconds
	// Default is 10000
	Timeout int

	// how many commands of the same type can run at the same time
	// Default is 10000
	MaxConcurrentRequests int

	// the minimum number of requests needed before a circuit can be tripped due to health
	// Default is 100
	RequestVolumeThreshold int

	//  how long, in milliseconds, to wait after a circuit opens before testing for recovery
	// Default is 500
	SleepWindow int

	// causes circuits to open once the rolling measure of errors exceeds this percent of requests
	// Default is 5
	ErrorPercentThreshold int
}

func (cb *CircuitBreakerConfig) init() {
	if cb.Timeout == 0 {
		cb.Timeout = defaultTimeout
	}

	if cb.MaxConcurrentRequests == 0 {
		cb.MaxConcurrentRequests = defaultMaxConcurrentRequests
	}

	if cb.RequestVolumeThreshold == 0 {
		cb.RequestVolumeThreshold = defaultRequestVolumeThreshold
	}

	if cb.SleepWindow == 0 {
		cb.SleepWindow = defaultSleepWindow
	}

	if cb.ErrorPercentThreshold == 0 {
		cb.ErrorPercentThreshold = defaultErrorPercentThreshold
	}
}

func (emq *EnqueueStomp) ConfigureCircuitBreaker(name string, cb CircuitBreakerConfig) {
	emq.mu.Lock()
	defer emq.mu.Unlock()

	cb.init()
	circuitName := emq.makeCircuitName(name)
	hystrix.ConfigureCommand(circuitName, hystrix.CommandConfig{
		Timeout:                cb.Timeout,
		MaxConcurrentRequests:  cb.MaxConcurrentRequests,
		RequestVolumeThreshold: cb.RequestVolumeThreshold,
		SleepWindow:            cb.SleepWindow,
		ErrorPercentThreshold:  cb.ErrorPercentThreshold,
	})
	emq.circuitNames[circuitName] = circuitName
}

func SetLogger() {
	hystrix.SetLogger(logger)
}

func (emq *EnqueueStomp) sendWithCircuitBreaker(circuitName string, destination string, body []byte, so SendOptions) error {
	err := hystrix.Do(circuitName, func() error {
		logger.Print(
			"[EnqueueStomp] send message with CircuitBreaker\n",
		)
		return emq.conn.Send(destination, so.ContentType, body, so.FrameOpts...)
	}, nil)

	return err
}

func (emq *EnqueueStomp) makeCircuitName(name string) string {
	return fmt.Sprintf("%s::%s", name, emq.id)
}

func (emq *EnqueueStomp) hasCircuitBreaker(name string) bool {
	_, found := emq.circuitNames[name]
	return found
}

//
