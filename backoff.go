/*
* enqueuestomp
*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */

package enqueuestomp

import (
	"math"
	"time"
)

// BackoffStrategy is used to determine how long a retry request should wait until attempted.
type BackoffStrategy func(retry int) time.Duration

const DefaultInitialBackOff = 100 * time.Millisecond

// ConstantBackOff always returns 100 Millisecond.
func ConstantBackOff(_ int) time.Duration {
	return DefaultInitialBackOff
}

// ExponentialBackoff returns ever increasing backoffs by a power of 2.
func ExponentialBackoff(i int) time.Duration {
	return time.Duration(math.Pow(2, float64(i))) * DefaultInitialBackOff // nolint:gomnd
}

// LinearBackoff returns increasing durations.
func LinearBackoff(i int) time.Duration {
	return time.Duration(i) * DefaultInitialBackOff
}
