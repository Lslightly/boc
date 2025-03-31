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

func assertWhenMergeSort(arr []int) {
	res := mergeSort(arr)
	slices.Sort(arr)
	if !slices.Equal(arr, res) {
		panic(fmt.Sprintf("%v != boc get %v", arr, res))
	}
}

func testWhenMergeSort(arr []int) {
	ch := make(chan bool)
	go func() {
		assertWhenMergeSort(arr)
		ch <- true
	}()
	<-ch
}

// TODO: not passed yet
func BenchmarkMergeSort(b *testing.B) {
	ch := make(chan bool)
	go func() {
		assertWhenMergeSort([]int{2, 3, 1, 4})

		arr2 := []int{3, 4, 2, 1, 8, 5, 6, 7}
		assertWhenMergeSort(arr2)
		slices.Sort(arr2)

		assertWhenMergeSort(arr2)

		arr3 := arr2
		arr3 = append(arr3, arr3...)
		arr3 = append(arr3, arr3...)
		arr3 = append(arr3, arr3...)
		assertWhenMergeSort(arr3)

		arr4 := make([]int, 0, 1024)
		for i := 1023; i >= 0; i-- {
			arr4 = append(arr4, i)
		}
		assertWhenMergeSort(arr4)

		ch <- true
	}()
	<-ch
}

// TODO: not passed yet
func TestMergeSort1(t *testing.T) {
	testWhenMergeSort([]int{3})
}

func TestMergeSort2(t *testing.T) {
	testWhenMergeSort([]int{3, 2})
}

// TODO: got stuck
// recover runtime error: slice bounds out of range [:5] with capacity 3
func TestMergeSort3(t *testing.T) {
	testWhenMergeSort([]int{3, 2, 1})
}

func TestMergeSort4(t *testing.T) {
	testWhenMergeSort([]int{3, 2, 1, 4})
}

func TestMergeSort5(t *testing.T) {
	testWhenMergeSort([]int{3, 2, 1, 4, 5})
}

func TestMergeSort6(t *testing.T) {
	testWhenMergeSort([]int{2, 3, 4, 8, 6, 10})
}

func TestMergeSort7(t *testing.T) {
	testWhenMergeSort([]int{3, 4, 2, 1, 8, 5, 6})
}

func TestMergeSort8(t *testing.T) {
	testWhenMergeSort([]int{3, 4, 2, 1, 8, 5, 6, 7})
}

type Step struct {
	start     int
	step_size int
	n         int
}

func (s Step) String() string {
	return fmt.Sprintf("(%d, %d)", s.start, s.step_size)
}

func (s Step) left_right() (left, right Step) {
	half := s.step_size / 2
	start := s.start
	left = Step{start, half, s.n}
	right = Step{start + half, half, s.n}
	return
}

func (s Step) debugRange() []int {
	r := make([]int, 0, s.step_size)
	for i := s.start; i < s.start+s.step_size && i < s.n; i++ {
		r = append(r, i)
	}
	return r
}

func (s Step) debugWaiting() {
	left, right := s.left_right()
	if right.start >= s.n {
		fmt.Println(s, "waiting for", left, "+", s.debugRange())
	} else {
		fmt.Println(s, "waiting for", left, right, "+", s.debugRange())
	}
}

func (s Step) release() string {
	left, right := s.left_right()
	return fmt.Sprintln(s, "release", left, right, "+", s.debugRange())
}

func mergeSortInner(start, step_size, n int, boc_arr []CownPtr[int], boc_map_finish map[Step]CownPtr[int], sender chan []int, step_ch chan Step) {
	curStep := Step{start, step_size, n}
	leftStep, rightStep := curStep.left_right()

	hasRight := true
	if rightStep.start >= n { // out of range
		hasRight = false
	}

	curStep.debugWaiting()

	end := min(start+step_size, n)
	bocs := boc_arr[start:end]

	if !hasRight {
		bocs = append(bocs, boc_map_finish[curStep], boc_map_finish[leftStep])
	} else {
		bocs = append(bocs, boc_map_finish[curStep], boc_map_finish[leftStep], boc_map_finish[rightStep])
	}
	WhenVec(bocs, func(arr_finish ...*int) {
		all_len := len(arr_finish)
		var cur_finish, left_finish, right_finish *int
		one := 1
		right_finish = &one
		if hasRight {
			cur_finish, left_finish, right_finish = arr_finish[all_len-3], arr_finish[all_len-2], arr_finish[all_len-1]
		} else {
			cur_finish, left_finish = arr_finish[all_len-2], arr_finish[all_len-1]
		}
		step_ch <- Step{start, step_size, n}
		// left and right not finished yet
		if *left_finish == 0 {
			fmt.Println(leftStep, "not finished yet", curStep.release())
			return
		}
		if *right_finish == 0 {
			fmt.Println(rightStep, "not finished yet", curStep.release())
			return
		}
		if *cur_finish == 1 { // cur_finish finished
			fmt.Println(curStep, "already finished", curStep.release())
			return
		}
		arrToBeSorted := arr_finish[:all_len-3]
		arrLen := len(arrToBeSorted)

		lo := 0
		hi := step_size / 2
		res := make([]int, 0, arrLen)
		for range arrLen {
			if lo >= step_size/2 {
				res = append(res, *arr_finish[hi])
				hi += 1
				continue
			}
			if hi >= arrLen {
				res = append(res, *arr_finish[lo])
				lo += 1
				continue
			}
			if *arr_finish[lo] > *arr_finish[hi] {
				res = append(res, *arr_finish[hi])
				hi += 1
				continue
			}
			res = append(res, *arr_finish[lo])
			lo += 1
		}
		if start == 0 && step_size >= n {
			sender <- res
			return
		}
		for i := range arrLen {
			*arr_finish[i] = res[i] // write back [start:end)
		}

		*cur_finish = 1
		fmt.Println(curStep, "finished", curStep.release())

		new_step_size := step_size * 2
		mergeSortInner(start/new_step_size*new_step_size, new_step_size, n, boc_arr, boc_map_finish, sender, step_ch)
	}, DefaultPostFn)
}

func mergeSort(array []int) []int {
	n := len(array)
	if n == 1 {
		return array
	}

	boc_arr := make([]CownPtr[int], n)
	for i, v := range array {
		boc_arr[i] = NewCownPtr(v)
	}
	boc_finish_map := make(map[Step]CownPtr[int])
	step_size := 1
	for step_size <= n {
		for i := 0; i < n; i += step_size {
			boc_finish_map[Step{i, step_size, n}] = NewCownPtr(0)
		}
		step_size *= 2
	}
	if step_size > n { // for finish
		boc_finish_map[Step{0, step_size, n}] = NewCownPtr(0)
	}
	finish_ch := make(chan []int)
	defer close(finish_ch)
	step_ch := make(chan Step)
	defer close(step_ch)

	for i := range n {
		step := Step{i, 1, n}
		When2(boc_arr[i], boc_finish_map[step], func(_ *int, finished *int) {
			*finished = 1
			fmt.Println(step, "finished")
			mergeSortInner(i/2*2, 2, n, boc_arr, boc_finish_map, finish_ch, step_ch)
		}, DefaultPostFn)
	}

	for {
		select {
		case <-step_ch:
		case res := <-finish_ch:
			return res
		default:
			continue
		}
	}
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
