package queue

// Queue represents a basic queue.
type Queue interface {
	// Enqueue adds an item to the end of the queue.
	Enqueue(item interface{}) error
	// Dequeue removes and returns the item from the front of the queue.
	Dequeue() (interface{}, error)
	// Size returns the current size of the queue.
	Size() (int, error)
	// ReadAllMessages reads all pending messages in the queue
	ReadAllMessages() ([]interface{}, error)
	// ClearQueue clears all messages from the queue.
	ClearQueue() error
}
