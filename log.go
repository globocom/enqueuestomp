/*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */
package enqueuestomp

import (
	"fmt"
	golog "log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger = golog.New(os.Stderr, "", golog.Ldate|golog.Ltime|golog.Lmicroseconds)

func Printf(identifier string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	logger.Printf("[EnqueueStomp][%s] %s", identifier, message)
}

func createOutputWriteOnDisk(outputPath string) (*zap.Logger, error) {
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		DisableCaller:    true,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		ErrorOutputPaths: []string{"stderr"},
		OutputPaths: []string{
			outputPath,
		},
	}
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	return config.Build()
}
