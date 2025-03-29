package boc

import (
	"reflect"
	"slices"
	"sync/atomic"
)

type behaviorThunk func()

type behavior struct {
	thunk    behaviorThunk
	count    atomic.Int64
	requests []*request
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
		thunk:    func() { f(cowns) },
		count:    atomic.Int64{},
		requests: rs,
	}
	b.count.Add(int64(len(rs) + 1))
	return b
}
func (b *behavior) schedule() {
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
		b.thunk()
		for _, r := range b.requests {
			r.release()
		}
	}()
}
