// queue package

package queue

import "sync"

const (
	// QueueBufferSize represents the maximum size of a queue
	QueueBufferSize = 1024
)

// InMemoryQueue implements an in-memory queue.
type InMemoryQueue struct {
	ch   chan interface{}
	lock sync.RWMutex
}

// NewInMemoryQueue creates a new queue.
func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		ch: make(chan interface{}, QueueBufferSize),
	}
}

// Enqueue adds an item to the end of the queue.
func (q *InMemoryQueue) Enqueue(item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.ch <- item
}

// Dequeue removes and returns the item from the front of the queue.
func (q *InMemoryQueue) Dequeue() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()
	return <-q.ch
}

// Size returns the current size of the queue.
func (q *InMemoryQueue) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.ch)
}

// ReadAllMessages reads all pending messages in the queue
func (q *InMemoryQueue) ReadAllMessages() []interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	var messages []interface{}
	for len(q.ch) > 0 {
		messages = append(messages, <-q.ch)
	}

	return messages
}

// ClearQueue clears all messages from the queue.
func (q *InMemoryQueue) ClearQueue() {
	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.ch) > 0 {
		<-q.ch
	}
}
