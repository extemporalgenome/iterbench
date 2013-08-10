package iter

import (
	"sort"
	"testing"
)

type (
	Item interface{}
	Iter func(more bool) (item Item, next Iter)
)

func IntSliceIter(s []int) Iter {
	if len(s) == 0 {
		return nil
	}
	var (
		f Iter
		i int
	)
	f = func(more bool) (item Item, next Iter) {
		item = s[i]
		i++
		if i < len(s) {
			return item, f
		}
		return item, nil
	}
	return f
}

func IntSliceCall(s []int, f func(Item) (more bool)) {
	for _, v := range s {
		if !f(v) {
			break
		}
	}
}

func IntSliceChan(s []int, done <-chan bool) <-chan Item {
	ch := make(chan Item)
	go func() {
		defer close(ch)
		for _, v := range s {
			select {
			case <-done:
				return
			case ch <- v:
			}
		}
	}()
	return ch
}

func IntKeyIter(m map[int]struct{}) Iter {
	if len(m) == 0 {
		return nil
	}
	var (
		f    Iter
		ch   = make(chan Item)
		done = make(chan bool)
	)
	go func() {
		defer close(ch)
		for k := range m {
			select {
			case <-done:
				return
			case ch <- k:
			}
		}
	}()
	lookahead := <-ch
	f = func(more bool) (Item, Iter) {
		if !more {
			close(done)
			return nil, nil
		}
		var item Item
		item, lookahead = lookahead, <-ch
		if lookahead == nil {
			return item, nil
		}
		return item, f
	}
	return f
}

func IntKeyCall(m map[int]struct{}, f func(Item) (more bool)) {
	for k := range m {
		if !f(k) {
			break
		}
	}
}

func IntKeyChan(m map[int]struct{}, done <-chan bool) <-chan Item {
	ch := make(chan Item)
	go func() {
		defer close(ch)
		for k := range m {
			select {
			case <-done:
				return
			case ch <- k:
			}
		}
	}()
	return ch
}

func Expand(next Iter) []Item {
	var (
		s []Item
		v Item
	)
	for next != nil {
		v, next = next(true)
		s = append(s, v)
	}
	return s
}

type intItemSlice []Item

func (s intItemSlice) Len() int           { return len(s) }
func (s intItemSlice) Less(i, j int) bool { return s[i].(int) < s[j].(int) }
func (s intItemSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

const n = 1 << 10

var (
	s []int
	m map[int]struct{}
)

func init() {
	s = make([]int, n)
	m = make(map[int]struct{}, n)
	for i := range s {
		v := n - i
		s[i] = v
		m[v] = struct{}{}
	}
}

func TestIntSliceIter(t *testing.T) {
	item, next := Item(nil), IntSliceIter(s)
	for i := 0; next != nil; i++ {
		item, next = next(true)
		if item != n-i {
			t.FailNow()
		}
	}
}

func TestIntKey(t *testing.T) {
	s := Expand(IntKeyIter(m))
	sort.Sort(sort.Reverse(intItemSlice(s)))
	for i, item := range s {
		if item != n-i {
			t.FailNow()
		}
	}
}

func BenchmarkIntSliceLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range s {
		}
	}
}

func BenchmarkIntSliceIter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		next := IntSliceIter(s)
		for next != nil {
			_, next = next(true)
		}
	}
}

func BenchmarkIntSliceCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IntSliceCall(s, func(Item) bool { return true })
	}
}

func BenchmarkIntSliceChan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range IntSliceChan(s, nil) {
		}
	}
}

func BenchmarkIntKeyLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range m {
		}
	}
}

func BenchmarkIntKeyIter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		next := IntKeyIter(m)
		for next != nil {
			_, next = next(true)
		}
	}
}

func BenchmarkIntKeyCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IntKeyCall(m, func(Item) bool { return true })
	}
}

func BenchmarkIntKeyChan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range IntKeyChan(m, nil) {
		}
	}
}
