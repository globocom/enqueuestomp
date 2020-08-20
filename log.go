/*
* enqueuestomp
*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */

package enqueuestomp

import (
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

func (emq *EnqueueStomp) newOutput() (err error) {
	if emq.config.WriteOutputPath == "" {
		return nil
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

	if emq.output, err = config.Build(); err == nil {
		emq.hasOutput = true
	}

	return err
}

func (emq *EnqueueStomp) writeOutput(action string, identifier string, destinationType string, destinationName string, body []byte) {
	if emq.hasOutput {
		emq.output.Info(action,
			zap.String("identifier", identifier),
			zap.String("destinationType", destinationType),
			zap.String("destinationName", destinationName),
			zap.ByteString("body", body),
		)
	}
}

func (emq *EnqueueStomp) debugLogger(template string, args ...interface{}) {
	emq.log.Debugf(template, args)
}

func (emq *EnqueueStomp) errorLogger(template string, args ...interface{}) {
	emq.log.Errorf(template, args)
}
