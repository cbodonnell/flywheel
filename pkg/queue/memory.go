package queue

import (
	"fmt"
	"sync"
)

type InMemoryQueue struct {
	items    []interface{}
	capacity int
	lock     sync.RWMutex
}

// NewInMemoryQueue creates a new queue.
func NewInMemoryQueue(capacity int) Queue {
	return &InMemoryQueue{
		items:    make([]interface{}, 0),
		capacity: capacity,
	}
}

func (q *InMemoryQueue) Enqueue(item interface{}) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.items) == q.capacity {
		return fmt.Errorf("queue is full")
	}

	q.items = append(q.items, item)
	return nil
}

func (q *InMemoryQueue) Dequeue() (interface{}, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.items) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

func (q *InMemoryQueue) Size() (int, error) {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.items), nil
}

func (q *InMemoryQueue) ReadAllMessages() ([]interface{}, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	messages := make([]interface{}, len(q.items))
	copy(messages, q.items)
	q.items = make([]interface{}, 0)
	return messages, nil
}

func (q *InMemoryQueue) ClearQueue() error {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items = make([]interface{}, 0)
	return nil
}
