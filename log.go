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

func (emq *EnqueueStompImpl) newOutput() (err error) {
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

func (emq *EnqueueStompImpl) writeOutput(action string, identifier string, destinationType string, destinationName string, body []byte, logField LogField) {
	if emq.hasOutput {
		fields := []zap.Field{
			zap.String("identifier", identifier),
			zap.String("destinationType", destinationType),
			zap.String("destinationName", destinationName),
			zap.ByteString("body", body),
		}

		if logField != nil && len(logField.getFields()) > 0 {
			fields = append(fields, logField.getFields()...)
		}

		emq.output.Info(action, fields...)
	}
}

func (emq *EnqueueStompImpl) debugLogger(template string, args ...interface{}) {
	emq.log.Debugf(template, args...)
}

func (emq *EnqueueStompImpl) errorLogger(template string, args ...interface{}) {
	emq.log.Errorf(template, args...)
}
