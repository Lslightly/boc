package boc

import (
	"fmt"
	"math/rand"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkMessagePassing(b *testing.B) {
	for b.Loop() {
		c1 := NewCownPtr(false)
		t1 := atomic.Uint64{}
		ch1 := make(chan int, 1)
		ch2 := make(chan int, 1)
		go func() {
			When1(c1, func(sent *bool) {
				if !*sent {
					t1.Add(1)
					*sent = true
				} else {
					if t1.Load() != 1 {
						panic("t1 should be 1")
					}
				}
				ch1 <- 1
			}, DefaultPostFn)
		}()
		go func() {
			When1(c1, func(sent *bool) {
				if !*sent {
					t1.Add(1)
					*sent = true
				} else {
					if t1.Load() != 1 {
						panic("t1 should be 1")
					}
				}
				ch2 <- 1
			}, DefaultPostFn)
		}()
		<-ch1
		<-ch2
		if t1.Load() != 1 {
			println(t1.Load())
			b.Fatalf("t1 should be 1")
		}
	}
}

func BenchmarkMessagePassingDeterminesOrder(b *testing.B) {
	for b.Loop() {
		cflag1 := NewCownPtr(false)
		cflag2 := NewCownPtr(false)
		t1 := atomic.Uint64{}

		ch := make(chan int)

		go func() {
			When1(cflag1, func(flag1 *bool) { *flag1 = true }, DefaultPostFn)

			if !t1.CompareAndSwap(0, 1) {
				When1(cflag2, func(flag2 *bool) {
					if !*flag2 {
						panic("flag2 should be true")
					}
					ch <- 1
				}, DefaultPostFn)
			}
		}()

		go func() {
			When1(cflag2, func(flag2 *bool) { *flag2 = true }, DefaultPostFn)

			if !t1.CompareAndSwap(0, 1) {
				When1(cflag1, func(flag1 *bool) {
					if !*flag1 {
						panic("flag2 should be true")
					}
					ch <- 1
				}, DefaultPostFn)
			}
		}()
		<-ch

		if t1.Load() != 1 {
			b.Fatalf("t1 should be 1")
		}
	}
}

func BenchmarkFibonacci(b *testing.B) {
	for b.Loop() {
		ch := make(chan int)
		go func() {
			accumulator := []int{0, 1}
			for i := 2; i <= 25; i++ {
				answer := accumulator[i-1] + accumulator[i-2]
				accumulator = append(accumulator, answer)
				bocAnswer := fibonacci(i)
				if answer != bocAnswer {
					panic(fmt.Sprintf("fibonacci(%d) should be %d, but boc get %d", i, answer, bocAnswer))
				}
			}

			ch <- 1
		}()
		<-ch
	}
}

func fibonacci_inner(n int, ch chan int) CownPtr[int] {
	if n == 0 {
		return NewCownPtr(0)
	} else if n == 1 {
		return NewCownPtr(1)
	} else {
		prev := fibonacci_inner(n-1, nil)
		pprev := fibonacci_inner(n-2, nil)
		When2(prev, pprev, func(g1, g2 *int) {
			*g1 += *g2
			if ch != nil {
				ch <- *g1
			}
		}, DefaultPostFn)
		return prev
	}
}

func fibonacci(n int) int {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	}

	ch := make(chan int)
	fibonacci_inner(n, ch)
	return <-ch
}

// TODO: not passed yet
func BenchmarkMergeSort(b *testing.B) {
	ch := make(chan bool)
	go func() {
		arr1 := []uint64{2, 3, 1, 4}
		res1 := mergeSort(arr1)
		slices.Sort(arr1)
		if !slices.Equal(arr1, res1) {
			panic(fmt.Sprintf("mergeSort(%v) should be %v, but boc get %v", arr1, arr1, res1))
		}

		arr2 := []uint64{3, 4, 2, 1, 8, 5, 6, 7}
		res2 := mergeSort(arr2)
		slices.Sort(arr2)
		if !slices.Equal(arr2, res2) {
			panic(fmt.Sprintf("mergeSort(%v) should be %v, but boc get %v", arr2, arr2, res2))
		}

		res2_ := mergeSort(arr2)
		if !slices.Equal(arr2, res2_) {
			panic(fmt.Sprintf("mergeSort(%v) should be %v, but boc get %v", arr2, arr2, res2_))
		}

		arr3 := arr2
		arr3 = append(arr3, arr3...)
		arr3 = append(arr3, arr3...)
		arr3 = append(arr3, arr3...)
		res3 := mergeSort(arr3)
		slices.Sort(arr3)
		if !slices.Equal(arr3, res3) {
			panic(fmt.Sprintf("mergeSort(%v) should be %v, but boc get %v", arr3, arr3, res3))
		}

		arr4 := make([]uint64, 0, 1024)
		for i := 1023; i >= 0; i-- {
			arr4 = append(arr4, uint64(i))
		}
		res4 := mergeSort(arr4)
		slices.Sort(arr4)
		if !slices.Equal(arr4, res4) {
			panic(fmt.Sprintf("mergeSort(%v) should be %v, but boc get %v", arr4, arr4, res4))
		}

		ch <- true
	}()
	<-ch
}

// TODO: not passed yet
func TestHardMergeSort(t *testing.T) {
	ch := make(chan bool)
	go func() {
		arr2 := []uint64{3, 4, 2, 1, 8, 5, 6, 7}
		res2 := mergeSort(arr2)
		slices.Sort(arr2)
		if !slices.Equal(arr2, res2) {
			panic(fmt.Sprintf("%v != boc get %v", arr2, res2))
		}
		ch <- true
	}()
	<-ch
}

func mergeSortInner(idx, step_size, n uint64, boc_arr []CownPtr[uint64], boc_finish []CownPtr[uint64], sender chan []uint64) {
	if idx == 0 {
		return
	}

	from := idx*step_size - n
	to := (idx+1)*step_size - n

	bocs := boc_arr[from:to]
	bocs = append(bocs, boc_finish[idx], boc_finish[idx*2], boc_finish[idx*2+1])

	WhenVec(bocs, func(content ...*uint64) {
		left_and_right_sorted := (*content[step_size+1] == 1) && (*content[step_size+2] == 1)
		if !left_and_right_sorted || *content[step_size] == 1 {
			// If both subarrays are not ready or we already sorted for this range, skip.
			return
		}

		lo := uint64(0)
		hi := uint64(step_size / 2)
		res := make([]uint64, 0)
		for uint64(len(res)) < step_size {
			if lo >= step_size/2 || (hi < step_size && *content[lo] > *content[hi]) {
				res = append(res, *content[hi])
				hi += 1
			} else {
				res = append(res, *content[lo])
				lo += 1
			}
		}
		for i := range step_size {
			*content[i] = res[i]
		}

		// Signal that we have sorted the subarray [from, to).
		*content[step_size] = 1

		// If the sorting process is completed send a signal to the main thread.
		if idx == 1 {
			sender <- res
			return
		}

		// Recursively sort the larger subarray (bottom up)
		mergeSortInner(idx/2, step_size*2, n, boc_arr, boc_finish, sender)
	}, DefaultPostFn)
}

func mergeSort(array []uint64) []uint64 {
	n := uint64(len(array))
	if n == 1 {
		return array
	}

	boc_arr := make([]CownPtr[uint64], n)
	for i, v := range array {
		boc_arr[i] = NewCownPtr(v)
	}
	boc_finish := make([]CownPtr[uint64], 2*n)
	for i := range 2 * n {
		boc_finish[i] = NewCownPtr(uint64(0))
	}

	finish_ch := make(chan []uint64)

	for i := range n {
		c_finished := boc_finish[i+n]
		When1(c_finished, func(finished *uint64) {
			// Signal that the sorting of subarray for [i, i+1) is finished.
			*finished = 1

			mergeSortInner(uint64((n+i)/2), 2, n, boc_arr, boc_finish, finish_ch)
		}, DefaultPostFn)
	}

	return <-finish_ch
}

func runTransactions(account_cnt, transaction_cnt uint64, use_sleep bool) {
	if account_cnt == 0 {
		panic("account_cnt should be greater than 0")
	}
	if transaction_cnt == 0 {
		panic("transaction_cnt should be greater than 0")
	}

	accouts := make([]CownPtr[uint64], account_cnt)
	for i := range account_cnt {
		accouts[i] = NewCownPtr(rand.Uint64())
	}
	c_remaining := NewCownPtr(transaction_cnt)

	finish_ch := make(chan bool)

	for range transaction_cnt {
		src := rand.Uint64() % account_cnt
		dst := rand.Uint64() % account_cnt
		if src == dst {
			dst = (dst + 1) % account_cnt
		}

		amount := rand.Uint64() % 2048
		random_sleep := use_sleep && rand.Uint64()%2 == 0

		c_src := accouts[src]
		c_dst := accouts[dst]

		When2(c_src, c_dst, func(src, dst *uint64) {
			if amount <= *src {
				*src -= amount
				*dst += amount
			}

			if random_sleep {
				time.Sleep(time.Second)
			}

			When1(c_remaining, func(remaining *uint64) {
				*remaining -= 1
				if *remaining == 0 {
					finish_ch <- true
				}
			}, DefaultPostFn)
		}, DefaultPostFn)
	}

	<-finish_ch
}

func BenchmarkBanking(b *testing.B) {
	ch := make(chan bool)
	go func() {
		runTransactions(20, 20, true)
		ch <- true
	}()
	<-ch
}
