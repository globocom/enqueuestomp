/*
* enqueuestomp
*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */

package enqueuestomp

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

// NoopLogger does not log anything.
type NoopLogger struct{}

// Debugf does nothing.
func (l NoopLogger) Debugf(template string, args ...interface{}) {}

// Errorf does nothing.
func (l NoopLogger) Errorf(template string, args ...interface{}) {}

func (emq *EnqueueStomp) createOutput() (err error) {
	if !emq.config.WriteOnDisk {
		return nil
	}

	if emq.config.WriteOutputPath == "" {
		emq.config.WriteOutputPath = fmt.Sprintf("enqueuestomp-%s.log", emq.id)
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		DisableCaller:    true,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		ErrorOutputPaths: []string{"stderr"},
		OutputPaths: []string{
			emq.config.WriteOutputPath,
		},
	}
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	emq.output, err = config.Build()
	return err
}

func (emq *EnqueueStomp) writeOutput(action string, identifier string, destinationType string, destinationName string, body []byte) {
	emq.output.Info(action,
		zap.String("identifier", identifier),
		zap.String("destinationType", destinationType),
		zap.String("destinationName", destinationName),
		zap.ByteString("body", body),
	)
}

func (emq *EnqueueStomp) debugLogger(template string, args ...interface{}) {
	emq.log.Debugf(template, args)
}

func (emq *EnqueueStomp) errorLogger(template string, args ...interface{}) {
	emq.log.Errorf(template, args)
}
