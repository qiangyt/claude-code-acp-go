package utils

import (
	"sync"
)

// PushableT 是可推送的异步可迭代对象的具体类型
type PushableT[T any] struct {
	ch     chan T
	closed bool
	mu     sync.RWMutex
}

// Pushable 是 PushableT 的指针类型别名
type Pushable[T any] = *PushableT[T]

// NewPushable 创建新的 Pushable
func NewPushable[T any](bufferSize int) Pushable[T] {
	return &PushableT[T]{
		ch: make(chan T, bufferSize),
	}
}

// Push 添加元素到流
func (p *PushableT[T]) Push(item T) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return // 已关闭，忽略
	}

	p.ch <- item
}

// End 结束流
func (p *PushableT[T]) End() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return // 已经关闭
	}

	p.closed = true
	close(p.ch)
}

// Channel 返回底层 channel
func (p *PushableT[T]) Channel() <-chan T {
	return p.ch
}

// IsClosed 检查流是否已关闭
func (p *PushableT[T]) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}
