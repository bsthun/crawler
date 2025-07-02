package main

import (
	"sync"
)

type Pool[T any] struct {
	objects chan T
	mu      sync.Mutex
}

func NewPool[T any](items []T) *Pool[T] {
	pool := &Pool[T]{
		objects: make(chan T, len(items)*2),
	}

	// * add all items to the pool
	for _, item := range items {
		pool.objects <- item
		pool.objects <- item
	}

	return pool
}

func (p *Pool[T]) Get() T {
	return <-p.objects
}

func (p *Pool[T]) Put(obj T) {
	select {
	case p.objects <- obj:
	default:
	}
}

func (p *Pool[T]) Size() int {
	return len(p.objects)
}
