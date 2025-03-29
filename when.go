package boc

import (
	"os"
	"sync"
)

func runWhen(f func(CownIfaceVec), cowns CownIfaceVec) {
	b := newBehavior(f, cowns)
	b.schedule()
}

// Use When1, When2, When3, When4 or WhenVec instead of this function if possible to leverage compiler type checking.
func When(cowns CownIfaceVec, f func(cowns CownIfaceVec)) {
	var realfn func(CownIfaceVec) = f
	if typecheck {
		wg.Add(1) // add 1 to wait group
		realfn = func(cpv CownIfaceVec) {
			defer wg.Done() // sync with main goroutine
			f(cpv)
		}
	}
	runWhen(realfn, cowns)
}

// DefaultPostFn is a default post function that does nothing.
var DefaultPostFn = func() {}

// When1 is a helper function for When with one cown.
// postFn is a function that will be called after fn.
// It is useful for cleanup and handling panic in spawned goroutine.
//
// See example_test.go and boc_test.go for usage.
//
// See ExampleTypeCheckHelper for usage of postFn to handle panic.
func When1[T any](cown CownPtr[T], fn func(*T), postFn func()) {
	When(CownIfaceVec{cown}, gen1(fn, postFn))
}

// When2 is a helper function for When with two cown.
// See documentation of When1 for further explanation.
func When2[T0, T1 any](cown0 CownPtr[T0], cown1 CownPtr[T1], fn func(*T0, *T1), postFn func()) {
	When(CownIfaceVec{cown0, cown1}, gen2(fn, postFn))
}

func When3[T0, T1, T2 any](cown0 CownPtr[T0], cown1 CownPtr[T1], cown2 CownPtr[T2], fn func(*T0, *T1, *T2), postFn func()) {
	When(CownIfaceVec{cown0, cown1, cown2}, gen3(fn, postFn))
}

func When4[T0, T1, T2, T3 any](cown0 CownPtr[T0], cown1 CownPtr[T1], cown2 CownPtr[T2], cown3 CownPtr[T3], fn func(*T0, *T1, *T2, *T3), postFn func()) {
	When(CownIfaceVec{cown0, cown1, cown2, cown3}, gen4(fn, postFn))
}

// WhenVec is a helper function for a slice of cowns with same type.
// See documentation of When1 for further explanation.
func WhenVec[T any](cowns CownPtrVec[T], fn func(content ...*T), postFn func()) {
	When(FromCownPtrVec(cowns), genVec(fn, postFn))
}

func envHasTypeCheck() bool {
	s := os.Getenv("TYPE_CHECK")
	return s != "" && s != "0" && s != "false"
}

var typecheck bool = envHasTypeCheck()
var wg sync.WaitGroup

// Use this function to skip other operations, remaining only When statements.
// See ExampleTypeCheckHelper for usage.
func TypeCheckHelper(fn func()) {
	if !typecheck {
		fn()
	}
}

// Insert this function to the end of your main goroutine to wait for all When statements(goroutines spawned).
func TypeCheckEnd() {
	if typecheck {
		wg.Wait()
	}
}

// EnableTypeCheck is a function to enable type checking.
// Pass "TYPE_CHECK=1" to the environment variable can also enable type checking.
func EnableTypeCheck() {
	typecheck = true
}

func gen1[T any](fn func(*T), postFn func()) func(CownIfaceVec) {
	return func(cowns CownIfaceVec) {
		defer postFn()
		fn(&AsCownPtr[T](cowns[0]).inner.value)
	}
}

func gen2[T0, T1 any](fn func(*T0, *T1), postFn func()) func(CownIfaceVec) {
	return func(cowns CownIfaceVec) {
		defer postFn()
		fn(
			&AsCownPtr[T0](cowns[0]).inner.value,
			&AsCownPtr[T1](cowns[1]).inner.value,
		)
	}
}

func gen3[T0 any, T1 any, T2 any](fn func(*T0, *T1, *T2), postFn func()) func(CownIfaceVec) {
	return func(cowns CownIfaceVec) {
		defer postFn()
		fn(
			&AsCownPtr[T0](cowns[0]).inner.value,
			&AsCownPtr[T1](cowns[1]).inner.value,
			&AsCownPtr[T2](cowns[2]).inner.value,
		)
	}
}

func gen4[T0 any, T1 any, T2 any, T3 any](fn func(*T0, *T1, *T2, *T3), postFn func()) func(CownIfaceVec) {
	return func(cowns CownIfaceVec) {
		defer postFn()
		fn(
			&AsCownPtr[T0](cowns[0]).inner.value,
			&AsCownPtr[T1](cowns[1]).inner.value,
			&AsCownPtr[T2](cowns[2]).inner.value,
			&AsCownPtr[T3](cowns[3]).inner.value,
		)
	}
}

func genVec[T any](fn func(content ...*T), postFn func()) func(CownIfaceVec) {
	return func(cowns CownIfaceVec) {
		defer postFn()
		args := make([]*T, len(cowns))
		for i, cown := range cowns {
			args[i] = &AsCownPtr[T](cown).inner.value
		}
		fn(args...)
	}
}
