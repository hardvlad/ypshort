package main

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool sync.Pool
}

func New[T Resettable](f func() T) *Pool[T] {
	p := &Pool[T]{}
	p.pool.New = func() any {
		return f()
	}
	return p
}

func (p *Pool[T]) Get() T {
	v := p.pool.Get()
	obj := v.(T)
	obj.Reset()
	return obj
}

func (p *Pool[T]) Put(obj T) {
	p.pool.Put(obj)
}
