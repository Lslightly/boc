package boc

import (
	"fmt"
	"reflect"
	"slices"
	"sync/atomic"
	"time"
)

type behaviorThunk func()

var glbBId atomic.Int64

type behavior struct {
	thunk    behaviorThunk
	count    atomic.Int64
	requests []*request

	// for debug use
	id int64
}

// String returns a string representation of the behavior
func (b *behavior) String() string {
	return fmt.Sprintf("b%d", b.id)
}

func newBehavior(f func(CownIfaceVec), cowns CownIfaceVec) *behavior {
	rs := cowns.requests()
	slices.SortFunc(rs, func(a, b *request) int {
		aPtr := reflect.ValueOf(a.target).Pointer()
		bPtr := reflect.ValueOf(b.target).Pointer()
		if aPtr < bPtr {
			return -1
		} else if aPtr > bPtr {
			return 1
		} else {
			return 0
		}
	})
	b := &behavior{
		thunk:    nil,
		count:    atomic.Int64{},
		requests: rs,
		id:       glbBId.Add(1),
	}
	b.thunk = func() {
		cowns.storeBehavior(b)
		f(cowns)
	}
	b.count.Add(int64(len(rs) + 1))
	return b
}
func (b *behavior) schedule() {
	if debug {
		fmt.Println("schedule", b, "with", len(b.requests), "requests", b.requests)
	}
	for _, r := range b.requests {
		r.startEnqueue(b)
	}
	for _, r := range b.requests {
		r.finishEnqueue()
	}
	b.resolveOne()
}
func (b *behavior) resolveOne() {
	if b.count.Add(-1) != 0 {
		return
	}

	go func() {
		before := time.Now()
		defer func() {
			if r := recover(); r != nil {
				if debug {
					fmt.Println("behavior thunk panic", b)
				}
			}
			for _, r := range b.requests {
				if debug {
					fmt.Println(b, "release", r)
				}
				r.release()
			}
			d := time.Since(before)
			if debug {
				fmt.Println("done", b, "time:", d)
			}
		}()
		if debug {
			fmt.Println("start", b)
		}
		b.thunk()
	}()
}
