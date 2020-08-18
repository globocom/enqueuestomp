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
)

var logger = golog.New(os.Stderr, "", golog.Ldate|golog.Ltime|golog.Lmicroseconds)

func Printf(identifier string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	// logger.Printf("[EnqueueStomp] identifier: `%s` - message: `%s`", identifier, message)
	logger.Printf("[EnqueueStomp][%s] %s", identifier, message)
}
