package iter

import (
	"sort"
	"testing"
)

const n = 16 << 10

type (
	Iter     func() (int, Iter)
	IterFunc func(int)
	IterCh   <-chan int
)

var (
	Slice = make([]int, n)
	Map   = make(map[int]struct{}, n)
)

func init() {
	for i := range Slice {
		Slice[i] = i
		Map[i] = struct{}{}
	}
}

func IntSliceIter(s []int) Iter {
	if len(s) == 0 {
		return nil
	}
	var (
		f Iter
		i int
	)
	f = func() (item int, next Iter) {
		item = s[i]
		i++
		if i < len(s) {
			return item, f
		}
		return item, nil
	}
	return f
}

func IntSliceCall(s []int, f IterFunc) {
	for _, v := range s {
		f(v)
	}
}

func IntSliceChan(s []int) IterCh {
	ch := make(chan int)
	go func() {
		for _, v := range s {
			ch <- v
		}
		close(ch)
	}()
	return ch
}

func IntKeyIter(m map[int]struct{}) Iter {
	if len(m) == 0 {
		return nil
	}
	ch := make(chan int)
	go func() {
		for k := range m {
			ch <- k
		}
		close(ch)
	}()
	var f Iter
	look, ok := <-ch
	f = func() (item int, next Iter) {
		item = look
		look, ok = <-ch
		if !ok {
			return item, nil
		}
		return item, f
	}
	return f
}

func IntKeyCall(m map[int]struct{}, f IterFunc) {
	for k := range m {
		f(k)
	}
}

func IntKeyChan(m map[int]struct{}) IterCh {
	ch := make(chan int)
	go func() {
		for k := range m {
			ch <- k
		}
		close(ch)
	}()
	return ch
}

func CheckKeySlice(s []int) bool {
	if len(s) != n {
		return false
	}
	sort.Ints(s)
	for i, v := range s {
		if i != v {
			return false
		}
	}
	return true
}

func TestIntSliceIter(t *testing.T) {
	v, next := 0, IntSliceIter(Slice)
	for i := 0; next != nil; i++ {
		v, next = next()
		if v != i {
			t.FailNow()
		}
	}
}

func TestIntSliceCall(t *testing.T) {
	i := 0
	IntSliceCall(Slice, func(v int) {
		if v != i {
			t.FailNow()
		}
		i++
	})
}

func TestIntSliceChan(t *testing.T) {
	i := 0
	for v := range IntSliceChan(Slice) {
		if v != i {
			t.FailNow()
		}
		i++
	}
}

func TestIntKeyIter(t *testing.T) {
	var (
		next = IntKeyIter(Map)
		s    []int
		v    int
	)
	for next != nil {
		v, next = next()
		s = append(s, v)
	}
	if !CheckKeySlice(s) {
		t.FailNow()
	}
}

func TestIntKeyCall(t *testing.T) {
	var s []int
	IntKeyCall(Map, func(v int) {
		s = append(s, v)
	})
	if !CheckKeySlice(s) {
		t.FailNow()
	}
}

func TestIntKeyChan(t *testing.T) {
	var s []int
	for v := range IntKeyChan(Map) {
		s = append(s, v)
	}
	if !CheckKeySlice(s) {
		t.FailNow()
	}
}

func BenchmarkIntSliceLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range Slice {
		}
	}
}

func BenchmarkIntSliceIter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		next := IntSliceIter(Slice)
		for next != nil {
			_, next = next()
		}
	}
}

func BenchmarkIntSliceCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IntSliceCall(Slice, func(int) {})
	}
}

func BenchmarkIntSliceChan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range IntSliceChan(Slice) {
		}
	}
}

func BenchmarkIntKeyLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range Map {
		}
	}
}

func BenchmarkIntKeyIter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		next := IntKeyIter(Map)
		for next != nil {
			_, next = next()
		}
	}
}

func BenchmarkIntKeyCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IntKeyCall(Map, func(int) {})
	}
}

func BenchmarkIntKeyChan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range IntKeyChan(Map) {
		}
	}
}
