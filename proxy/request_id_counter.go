package proxy

import (
	"sync/atomic"
)

type RequestIDCounter struct {
	I uint64
}

func (r *RequestIDCounter) Next() uint64 {
	return atomic.AddUint64(&r.I, 1)
}
