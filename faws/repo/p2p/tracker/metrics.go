package tracker

import "sync/atomic"

type MeteredInt64 struct {
	value int64
}

func (metered_int64 *MeteredInt64) Add(delta int64) (new int64) {
	new = atomic.AddInt64(&metered_int64.value, delta)
	return
}

func (metered_int64 *MeteredInt64) Load() (i int64) {
	i = atomic.LoadInt64(&metered_int64.value)
	return
}

type ServerMetrics struct {
	TotalUsers   MeteredInt64
	CurrentUsers MeteredInt64
}

// func (m *Metrics) Log
