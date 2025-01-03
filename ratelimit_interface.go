package ratelimit

import (
	"time"
)

type BucketI interface {
	Wait(count int64)

	WaitMaxDuration(count int64, maxWait time.Duration) bool

	Take(count int64) time.Duration

	TakeMaxDuration(count int64, maxWait time.Duration) (time.Duration, bool)

	TakeAvailable(count int64) int64

	Available() int64

	Capacity() int64

	Rate() float64

	lock()

	unlock()

	take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool)

	takeAvailable(now time.Time, count int64) int64

	resetTokens(tokens int64)

	available(now time.Time) int64

	storeTokens()

	rollbackTokens()

	AddUpstream(bs ...BucketI)

	//ToBucketI() *BucketI
}
