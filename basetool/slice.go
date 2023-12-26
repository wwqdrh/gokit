package basetool

import (
	"errors"
	"sync/atomic"
)

type SafeSlice[T any] struct {
	head   *sliceNode[T]
	tail   *sliceNode[T]
	length int64
	ch     chan T // 通道用来接收添加数据
}

type sliceNode[T any] struct {
	Val  T
	Next *sliceNode[T]
}

func (SafeSlice[T]) NewSafeSlice(maxLength int) *SafeSlice[T] {
	s := &SafeSlice[T]{
		head:   nil,
		tail:   nil,
		length: 0,
		ch:     make(chan T, maxLength),
	}
	go func() {
		for item := range s.ch {
			cur := &sliceNode[T]{
				Val: item,
			}
			if atomic.LoadInt64(&s.length) == 0 {
				s.head = cur
				s.tail = cur
			} else {
				s.tail.Next = cur
				s.tail = cur
			}
			atomic.AddInt64(&s.length, 1)
		}
	}()

	return s
}

// 基于channel
func (s *SafeSlice[T]) Add(value T) {
	s.ch <- value
}

func (s *SafeSlice[T]) Get() ([]T, error) {
	length := atomic.LoadInt64(&s.length)
	if length == 0 {
		return nil, errors.New("数据为空")
	}

	data := []T{}

	var cur *sliceNode[T]
	for i := 0; i < int(length); i++ {
		if i == 0 {
			cur = s.head
		} else {
			cur = cur.Next
		}
		data = append(data, cur.Val)
	}
	return data, nil
}
