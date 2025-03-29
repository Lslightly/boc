package boc

import (
	"fmt"
	"strings"
)

func ExampleWhen2() {
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
			fmt.Println(*g1, *g2, *g3)
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
	// Output:
	// 2 5 true
}

func ExampleWhenVec() {
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
			fmt.Println(*x[0], *x[1], *g3)
			ch <- true
		}, DefaultPostFn)
	}, DefaultPostFn)
	<-ch
	// Output:
	// 1 2 true
}

func ExampleWhen() {
	c1 := NewCownPtr(0)
	c2 := NewCownPtr(0)
	c3 := NewCownPtr(false)
	ch := make(chan bool)
	When(CownIfaceVec{c1, c2}, func(cowns CownIfaceVec) {
		x0 := AsCownPtr[int](cowns[0]).AddrOfValue() // dynamic type checking
		x1 := AsCownPtr[int](cowns[1]).AddrOfValue()
		*x0 += 1
		*x1 += 1
		fmt.Println(*x0, *x1)
		When(CownIfaceVec{c3, c2}, func(cowns CownIfaceVec) {
			g3 := AsCownPtr[bool](cowns[0]).AddrOfValue()
			g2 := AsCownPtr[int](cowns[1]).AddrOfValue()
			*g2 += 1
			*g3 = true
			fmt.Println(*g2, *g3)
			ch <- true
		})
	})
	<-ch
	// Output:
	// 1 1
	// 2 true
}

func ExampleTypeCheckHelper() {
	EnableTypeCheck()
	c1 := NewCownPtr(1)
	c2 := NewCownPtr(false)

	t1 := CownIfaceVec{c1, c2}

	ch := make(chan bool)

	When(t1,
		gen2(
			func(g1 *bool, g2 *int) { // here g1, g2's type is wrong, should panic
				TypeCheckHelper(func() {
					*g1 = false
					*g2 += 2
					ch <- true
				})
			},
			func() {
				if r := recover(); r != nil {
					if strings.HasPrefix(r.(error).Error(), "convertion from interface") {
						fmt.Println(r)
						return
					}
				}
				panic("error is not interface convertion error")
			},
		),
	)
	TypeCheckHelper(func() {
		<-ch
	})
	TypeCheckWait()
	// Output:
	// convertion from interface{boc.CownPtr[int]} to bool failed
}
