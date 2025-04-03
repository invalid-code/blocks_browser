package main

import "iter"

type Stack[T any] struct {
	items []T
}

func (stack *Stack[T]) Push(items ...T) {
	for _, item := range items {
		stack.items = append(stack.items, item)
	}
}

func (stack *Stack[T]) Pop() T {
	item := stack.items[len(stack.items)-1]
	stack.items = stack.items[:len(stack.items)-1]
	return item
}

func (stack *Stack[T]) IsEmpty() bool {
	return len(stack.items) == 0
}

type Queue[T any] struct {
	items []T
}

func (queue *Queue[T]) Enqueue(items ...T) {
	for _, item := range items {
		queue.items = append(queue.items, item)
	}
}

func (queue *Queue[T]) Dequeue() T {
	item := queue.items[0]
	if len(queue.items) > 1 {
		queue.items = queue.items[1:]
	} else {
		queue.items = []T{}
	}
	return item
}

func (queue *Queue[T]) IsEmpty() bool {
	return len(queue.items) == 0
}

func convSeqToSlice[T any](seq iter.Seq[T]) []T {
	sliceToRet := []T{}
	for item := range seq {
		sliceToRet = append(sliceToRet, item)
	}
	return sliceToRet
}

func revSlice[T any](sliceToRev []T) []T {
	revedSlice := []T{}
	for i := len(sliceToRev) - 1; i > -1; i-- {
		revedSlice = append(revedSlice, sliceToRev[i])
	}
	return revedSlice
}
