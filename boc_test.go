package boc

import (
	"testing"
)

func TestBoc(t *testing.T) {
	c1 := NewCownPtr(1)
	c2 := NewCownPtr(2)
	c3 := NewCownPtr(false)

	ch := make(chan int)

	When2(c1, c2, func(g1 *int, g2 *int) {
		*g1 += 1
		*g2 += 2
		When2(c2, c3, func(g2 *int, g3 *bool) {
			*g2 += 1
			*g3 = true
		}, DefaultPostFn)
	}, DefaultPostFn)

	When3(c1, c2, c3, func(g1 *int, g2 *int, g3 *bool) {
		if *g1 != 2 {
			panic("g1 should be 2") // Attention! don't use t.Fatalf here
		}
		if *g2 != 4 {
			panic("g2 should be 4")
		}
		ch <- 1
	}, DefaultPostFn)

	<-ch
}

func TestBocSimple(t *testing.T) {
	c1 := NewCownPtr(1)
	c2 := NewCownPtr(2)
	c3 := NewCownPtr(false)

	When2(c1, c2, func(g1 *int, g2 *int) {
		*g1 += 1
		*g2 += 2
		When2(c3, c2, func(g3 *bool, g2 *int) {
			*g2 += 1
			*g3 = true
		}, DefaultPostFn)
	}, DefaultPostFn)
}

func TestBocChannel(t *testing.T) {
	c1 := NewCownPtr(1)
	c2 := NewCownPtr(2)
	c3 := NewCownPtr(false)

	ch := make(chan int)

	When2(c1, c2, func(g1 *int, g2 *int) {
		*g1 += 1
		*g2 += 2
		When2(c3, c2, func(g3 *bool, g2 *int) {
			*g2 += 1
			*g3 = true
			ch <- 1
		}, DefaultPostFn)
	}, DefaultPostFn)

	if <-ch != 1 {
		t.Fatalf("channel should be 1")
	}
}

func TestBocVec(t *testing.T) {
	c1 := NewCownPtr(0)
	c2 := NewCownPtr(0)
	c3 := NewCownPtr(false)

	ch := make(chan bool)
	WhenVec(CownPtrVec[int]{c1, c2}, func(x ...*int) {
		*x[0] += 1
		*x[1] += 1
		When2(c3, c2, func(g3 *bool, g2 *int) {
			*g2 += 1
			*g3 = true
		}, DefaultPostFn)
	}, DefaultPostFn)
	When3(c1, c2, c3, func(g1 *int, g2 *int, g3 *bool) {
		if *g1 != 1 {
			panic("g1 should be 1") // Attention! don't use t.Fatalf here
		}
		if *g3 == true {
			if *g2 != 2 {
				panic("g2 should be 2")
			}
		} else {
			if *g2 != 1 {
				panic("g2 should be 1")
			}
		}
		ch <- true
	}, DefaultPostFn)
	<-ch
}
