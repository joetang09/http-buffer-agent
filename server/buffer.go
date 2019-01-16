package server

import (
	"sync"
	"time"
)

// Buffer buffer of request
type Buffer struct {
	cap    int
	rear   int
	front  int
	mux    sync.Mutex
	buffer []interface{}
}

// NewBuffer buffer new function
func NewBuffer(cap int) *Buffer {
	return &Buffer{
		cap:    cap,
		rear:   0,
		front:  0,
		buffer: make([]interface{}, cap),
	}
}

// Put put element
func (b *Buffer) Put(i interface{}, timeout time.Time) bool {
	for {
		b.mux.Lock()
		if (b.rear+1)%b.cap != b.front {
			b.buffer[b.rear] = i
			b.rear = (b.rear + 1) % b.cap
			b.mux.Unlock()
			break
		}
		b.mux.Unlock()
		if time.Now().After(timeout) {
			return false
		}
	}
	return true

}

// Get get element
func (b *Buffer) Get(timeout time.Time) (interface{}, bool) {
	var r interface{}
	for {
		b.mux.Lock()
		if b.rear != b.front {
			r = b.buffer[b.front]
			b.front = (b.front + 1) % b.cap
			b.mux.Unlock()
			break
		}
		b.mux.Unlock()
		if time.Now().After(timeout) {
			return nil, false
		}
	}
	return r, true
}

// Empty buffer is empty or not
func (b *Buffer) Empty() bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.rear == b.front
}

// Clear buffer clear
func (b *Buffer) Clear() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.rear = 0
	b.front = 0
}

// Len buffer length
func (b *Buffer) Len() int {
	b.mux.Lock()
	defer b.mux.Unlock()
	return (b.rear - b.front + b.cap) % b.cap
}
