package ratelimit

import (
	"math"
	"time"
)

type ControlModel int

const (
	// SerialControl 串行控制
	// 使用串行控制时，只有所有上游桶的令牌数大于等于请求的令牌数时，才会从上游桶中取令牌
	SerialControl ControlModel = iota
	// ParallelControl 并行控制
	// 使用并行控制时，只要有一个上游桶的令牌数大于等于请求的令牌数时，就会从上游桶中取令牌
	ParallelControl
)

type BucketPlus struct {
	this           interface{ BucketI }
	upstreamBucket []interface{ BucketI }
	// 支持上游的并行控制或串行控制
	controlModel ControlModel
}

func NewBucketPlusN(upstream interface{ BucketI }, bs ...interface{ BucketI }) {
	//buckets := make([]*BucketPlus, 0, len(bs))
	for _, b := range bs {
		NewBucketPlus(SerialControl, upstream, b)

		//buckets = append(buckets, NewBucketPlus(SerialControl, upstream, b))
		//return buckets
	}
}

func NewBucketPlus(model ControlModel, upstream interface{ BucketI }, b interface{ BucketI }) *BucketPlus {
	m := make([]interface{ BucketI }, 0, 1)
	plus := BucketPlus{
		this:           b,
		upstreamBucket: append(m, upstream),
		controlModel:   model,
	}
	return &plus
}

func (bp *BucketPlus) AddUpstream(bs ...interface{ BucketI }) {
	bp.addUpstream(bs)
}

func (bp *BucketPlus) addUpstream(bs ...interface{}) {
	for _, b := range bs {
		if tb, ok := b.(interface{ BucketI }); ok {
			bp.upstreamBucket = append(bp.upstreamBucket, tb)
		}
	}
}

func (tb *Bucket) AddUpstream(bs ...interface{ BucketI }) {
	tb.addUpstream(bs)
}

func (tb *Bucket) addUpstream(bs ...interface{}) {
	for index, b := range bs {
		print(index)
		if t, ok := b.(interface{ BucketI }); ok {
			if tb.upstreamBucket == nil {
				tb.upstreamBucket = NewBucketPlus(SerialControl, t, tb)
			} else {
				tb.upstreamBucket.addUpstream(&t)
			}
		}
	}
}

func (bp *BucketPlus) SetControlModel(model ControlModel) {
	bp.controlModel = model
}

func (bp *BucketPlus) Wait(count int64) {
	if d := bp.Take(count); d > 0 {
		time.Sleep(d)
	}
}

func (bp *BucketPlus) WaitMaxDuration(count int64, maxWait time.Duration) bool {
	d, ok := bp.TakeMaxDuration(count, maxWait)
	if d > 0 {
		time.Sleep(d)
	}
	return ok
}

func (bp *BucketPlus) Take(count int64) time.Duration {
	bp.lock()
	defer bp.unlock()
	now := time.Now()
	take, _ := bp.take(now, count, infinityDuration)
	return take
}

func (bp *BucketPlus) TakeMaxDuration(count int64, maxWait time.Duration) (time.Duration, bool) {
	bp.lock()
	defer bp.unlock()
	return bp.take(time.Now(), count, maxWait)
}

func (bp *BucketPlus) TakeAvailable(count int64) int64 {
	bp.lock()
	defer bp.unlock()
	return bp.takeAvailable(time.Now(), count)
}

func (bp *BucketPlus) takeAvailable(now time.Time, count int64) int64 {
	availableTokens := bp.this.available(now)
	count = bp.this.takeAvailable(now, count)
	available := bp.takeAvailableUpstream(now, count)
	if available < count {
		bp.resetTokens(availableTokens + count - available)
		return available
	}
	return count
}

func (bp *BucketPlus) Available() int64 {
	bp.lock()
	defer bp.unlock()
	return bp.available(time.Now())
}

func (bp *BucketPlus) available(now time.Time) int64 {
	available := bp.this.available(now)
	available2 := bp.availableUpstream(now)
	if available2 < available {
		return available2
	}
	return available
}

func (bp *BucketPlus) Capacity() int64 {
	return bp.this.Capacity()
}

func (bp *BucketPlus) Rate() float64 {
	return bp.this.Rate()
}

func (bp *BucketPlus) lock() {
	bp.this.lock()
	for _, bs := range bp.upstreamBucket {
		bs.lock()
	}
}

func (bp *BucketPlus) unlock() {
	bp.this.unlock()
	for _, bs := range bp.upstreamBucket {
		bs.lock()
	}
}

func (bp *BucketPlus) resetTokens(tokens int64) {
	bp.this.resetTokens(tokens)
}

func (bp *BucketPlus) take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	if count <= 0 {
		return 0, true
	}

	//tokens := bp.available(now)
	bp.this.storeTokens()
	take, b2 := bp.this.take(now, count, maxWait)
	if !b2 {
		return 0, false
	}

	take2, b2 := bp.takeUpstream(now, count, maxWait)
	if !b2 {
		bp.this.rollbackTokens()
		return 0, false
	}

	if take2 > take {
		return take2, true
	}
	return take, true
}

func (tb *Bucket) storeTokens() {
	tb.lastAvailableTokens = tb.availableTokens
}

func (tb *Bucket) rollbackTokens() {
	tb.availableTokens = tb.lastAvailableTokens
}

func (bp *BucketPlus) storeTokens() {
	for _, bs := range bp.upstreamBucket {
		bs.storeTokens()
	}
}

func (bp *BucketPlus) rollbackTokens() {
	for _, bs := range bp.upstreamBucket {
		bs.rollbackTokens()
	}
}

func (bp *BucketPlus) takeUpstream(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	bp.storeTokens()
	upstream, b2 := bp._takeUpstream(now, count, maxWait)
	if !b2 {
		bp.rollbackTokens()
		return 0, false
	}
	return upstream, true
}

func (bp *BucketPlus) _takeUpstream(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	var take time.Duration = -1
	for _, bs := range bp.upstreamBucket {
		var t time.Duration
		var bo bool
		if bp, ok := bs.(*BucketPlus); ok {
			t, bo = bp._takeUpstream(now, count, maxWait)
		} else {
			t, bo = bs.take(now, count, maxWait)
		}
		if bp.controlModel == SerialControl && !bo {
			return 0, false
		}
		if bp.controlModel == ParallelControl && bo {
			return t, true
		}
		if bo && t > take {
			take = t
		}
	}
	if take == -1 {
		return 0, false
	}
	return take, true
}

func (bp *BucketPlus) availableUpstream(now time.Time) int64 {
	var available int64 = math.MaxInt64
	for _, bs := range bp.upstreamBucket {
		available2 := bs.available(now)
		if available2 < available {
			available = available2
		}
	}
	return available
}

func (bp *BucketPlus) takeAvailableUpstream(now time.Time, count int64) int64 {
	bp.storeTokens()
	available := bp._takeAvailableUpstream(now, count)
	if available < count {
		bp.rollbackTokens()
		if available > 0 {
			bp._takeAvailableUpstream(now, available)
		}
	}
	return available
}

func (bp *BucketPlus) _takeAvailableUpstream(now time.Time, count int64) int64 {
	for _, bs := range bp.upstreamBucket {
		var take int64
		if bp, ok := bs.(*BucketPlus); ok {
			take = bp._takeAvailableUpstream(now, count)
		} else {
			take = bs.takeAvailable(now, count)
		}
		if take < count {
			count = take
		}
	}
	return count
}
