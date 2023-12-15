// queue package

package queue

import "sync"

const (
	// QueueBufferSize represents the maximum size of a queue
	QueueBufferSize = 1024
)

// MemoryQueue represents a basic queue.
type MemoryQueue struct {
	ch   chan interface{}
	lock sync.RWMutex
}

// NewMemoryQueue creates a new queue.
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		ch: make(chan interface{}, QueueBufferSize),
	}
}

// Enqueue adds an item to the end of the queue.
func (q *MemoryQueue) Enqueue(item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.ch <- item
}

// Dequeue removes and returns the item from the front of the queue.
func (q *MemoryQueue) Dequeue() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()
	return <-q.ch
}

// Size returns the current size of the queue.
func (q *MemoryQueue) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.ch)
}

// ReadAllMessages reads all pending messages in the queue
func (q *MemoryQueue) ReadAllMessages() []interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	var messages []interface{}
	for len(q.ch) > 0 {
		messages = append(messages, <-q.ch)
	}

	return messages
}

// ClearQueue clears all messages from the queue.
func (q *MemoryQueue) ClearQueue() {
	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.ch) > 0 {
		<-q.ch
	}
}
