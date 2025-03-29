package boc

import (
	"math/rand"
	"slices"
	"testing"
)

func BenchmarkStressfibonacci(b *testing.B) {
	for b.Loop() {
		ch := make(chan bool)
		go func() {
			if fibonacci(28) != 317811 {
				panic("fibonacci(28) should be 317811")
			}
			ch <- true
		}()
		<-ch
	}
}

func BenchmarkStressBanking(b *testing.B) {
	ch := make(chan bool)
	go func() {
		runTransactions(1234, 100000, false)
		ch <- true
	}()
	<-ch
}

func BenchmarkStressMergeSort(b *testing.B) {
	const ITER int = 4
	const LOGSZ_LO int = 10
	const LOGSZ_HI int = 13

	chs := make([]chan bool, ITER)
	for i, ch := range chs {
		logsize := LOGSZ_LO + i%(LOGSZ_HI-LOGSZ_LO)
		le := 1 << logsize
		arr := make([]uint64, le)
		for j := range le {
			arr[j] = rand.Uint64()
		}

		go func() {
			res := mergeSort(arr)
			slices.Sort(arr)
			if !slices.Equal(arr, res) {
				panic("sort failed")
			}
			ch <- true
		}()
	}
	for _, ch := range chs {
		<-ch
	}
}
