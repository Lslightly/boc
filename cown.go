package boc

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

type cownBase interface {
	getLast() *atomic.Pointer[request]
}

type cown[T any] struct {
	last  atomic.Pointer[request]
	value T
}

func (c *cown[T]) getLast() *atomic.Pointer[request] {
	return &c.last
}

// CownPtr is a wrapper for a cown[T].
//
// Use [NewCownPtr] to create a new CownPtr.
// User should not care about the inner data.
type CownPtr[T any] struct {
	inner *cown[T]
}

// [NewCownPtr] creates a new [CownPtr] with the given value.
// The value is stored in a cown, which allows for safe concurrent access.
// The [CownPtr] can be used to pass the value to other goroutines safely.
func NewCownPtr[T any](value T) CownPtr[T] {
	return CownPtr[T]{
		inner: &cown[T]{
			last:  atomic.Pointer[request]{},
			value: value,
		},
	}
}

// [CownPtr.AddrOfValue] returns the address of the value inside the [CownPtr].
// This is useful for passing the value to a function that requires a pointer.
func (ptr CownPtr[T]) AddrOfValue() *T {
	return &ptr.inner.value
}

type cownIface interface {
	requests() []*request
}

// AsCownPtr convert from cownIface to CownPtr[T].
// It panics if the conversion fails.
func AsCownPtr[T any](ptr cownIface) CownPtr[T] {
	if cown, ok := ptr.(CownPtr[T]); ok {
		return cown
	}
	var zero T
	panic(fmt.Errorf("convertion from interface{%s} to %T failed", reflect.TypeOf(ptr).String(), zero))
}

func (ptr CownPtr[T]) requests() []*request {
	return []*request{newRequest(ptr.inner)}
}

// [CownIfaceVec] is a slice of cownIface.
// cownIface can store different types of [CownPtr][T].
// The use case is [CownIfaceVec]{CownPtr[int], CownPtr[bool], ...}.
// But the type checking have to be deferred to runtime.
//
// If macro is supported, finite tuple type (CownPtr[int], (CownPtr[bool], ...) can replace this
type CownIfaceVec []cownIface

// [CownPtrVec] is a slice of [CownPtr] with same type parameter T.
// Use [CownPtrVec.ToIfaceVec] to convert it to [CownIfaceVec].
type CownPtrVec[T any] []CownPtr[T]

// [CownPtrVec.ToIfaceVec] converts a slice of [CownPtr][T] to [CownIfaceVec].
// This is useful for passing a slice of CownPtr[T] to a function that
// expects a [CownIfaceVec].
func (slice CownPtrVec[T]) ToIfaceVec() CownIfaceVec {
	vec := make(CownIfaceVec, len(slice))
	for i, cown := range slice {
		vec[i] = cown
	}
	return vec
}

func (vec CownIfaceVec) requests() []*request {
	var requests []*request
	for _, cown := range vec {
		requests = append(requests, cown.requests()...)
	}
	return requests
}
