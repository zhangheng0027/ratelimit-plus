package ratelimit

import (
	"sync"
	"time"
)

type usageLog struct {
	log         []int64
	lastSeconds int64
	rightIndex  int
	mu          sync.Mutex
	maxLen      int
}

func newUsageLog(len int) *usageLog {
	return &usageLog{
		log:         make([]int64, len),
		lastSeconds: time.Now().Unix(),
		rightIndex:  0,
		maxLen:      len,
	}
}

func (log *usageLog) addLog(t time.Time, d time.Duration, count int64) {
	if count <= 0 {
		return
	}
	log.mu.Lock()
	defer log.mu.Unlock()
	seconds := t.Unix()
	sub := seconds - log.lastSeconds
	if sub < 0 {
		return
	}
	if sub != 0 {
		for i := 0; i < int(sub) && i < log.maxLen; i++ {
			log.rightIndex = (log.rightIndex + 1) % log.maxLen
			log.log[log.rightIndex] = 0
		}
	}

	ts := int64(d / time.Second)
	log.lastSeconds = seconds + ts
	if ts == 0 {
		log.log[log.rightIndex] += count
		return
	}

	sp := int64(float64(count) / d.Seconds())
	for i := 0; i < int(ts) && i < log.maxLen; i++ {
		log.log[log.rightIndex] += sp
		log.rightIndex = (log.rightIndex + 1) % log.maxLen
	}
	log.rightIndex = (log.rightIndex - 1 + log.maxLen) % log.maxLen
}

func (log *usageLog) getRate(now time.Time) int64 {
	t := now.Unix()
	if t < log.lastSeconds {
		return 0
	}
	d := t - log.lastSeconds
	if d > 5 {
		return 0
	}

	d = int64(log.maxLen) - d

	index := log.rightIndex
	var sum int64 = 0
	for i := 0; i < int(d); i++ {
		sum += log.log[index]
		if index == 0 {
			index = log.maxLen
		}
		index--
	}
	return sum / 5
}
