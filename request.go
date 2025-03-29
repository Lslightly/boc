package boc

import (
	"sync/atomic"
)

// request for cown
type request struct {
	next      atomic.Pointer[behavior] // next behavior to be executed
	scheduled atomic.Bool              // indicates if the request has been scheduled
	target    cownBase                 // target cown the request wants to access
}

func newRequest(target cownBase) *request {
	return &request{
		next:      atomic.Pointer[behavior]{},
		scheduled: atomic.Bool{},
		target:    target,
	}
}

func (r *request) startEnqueue(behavior *behavior) {
	prev := r.target.getLast().Swap(r)
	if prev != nil {
		for !prev.scheduled.Load() {
		}
		prev.next.Store(behavior)
		return
	}
	behavior.resolveOne()
}

func (r *request) finishEnqueue() {
	r.scheduled.Store(true)
}

func (r *request) release() {
	if r.next.Load() == nil {
		if r.target.getLast().CompareAndSwap(r, nil) {
			return
		}
		for r.next.Load() == nil {
		}
	}
	r.next.Load().resolveOne()
}
