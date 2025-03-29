package boc

import (
	"fmt"
	"sync/atomic"
)

var glbReqId atomic.Int64

// request for cown
type request struct {
	next      atomic.Pointer[behavior] // next behavior to be executed
	scheduled atomic.Bool              // indicates if the request has been scheduled
	target    cownBase                 // target cown the request wants to access

	// for debug use
	prev_req *request // previous request in the queue
	id       int64    // request id for debug use
	done     bool     // request has been released
}

// String returns a string representation of the request
func (r *request) String() string {
	return fmt.Sprintf("r%d", r.id)
}

func newRequest(target cownBase) *request {
	return &request{
		prev_req:  nil,
		next:      atomic.Pointer[behavior]{},
		scheduled: atomic.Bool{},
		target:    target,
		id:        glbReqId.Add(1),
		done:      false,
	}
}

func (r *request) startEnqueue(behavior *behavior) {
	prev := r.target.getLast().Swap(r)
	if prev != nil {
		r.prev_req = prev
		first := true
		for !prev.scheduled.Load() {
			if first {
				first = false
				fmt.Println(r, "waiting", prev, "to be scheduled")
			}
		}
		if !first {
			fmt.Println(r, "waiting succ")
		}
		prev.next.Store(behavior)
		return
	}
	debugRequestResolveBehavior(r, behavior)
	behavior.resolveOne()
}

func (r *request) finishEnqueue() {
	r.scheduled.Store(true)
	fmt.Println("scheduled", r)
}

func (r *request) release() {
	if r.next.Load() == nil {
		if r.target.getLast().CompareAndSwap(r, nil) {
			return
		}
		for r.next.Load() == nil {
			fmt.Println(r, "'s next is nil, waiting for it to be set")
		}
	}
	b := r.next.Load()
	debugRequestResolveBehavior(r, b)
	b.resolveOne()
	r.done = true
}

func debugRequestResolveBehavior(r *request, b *behavior) {
	fmt.Println(r, ": --", b, "remaining", b.count.Load()-1, "/total", len(b.requests))
}
