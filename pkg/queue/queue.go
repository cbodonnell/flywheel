package queue

type Queue interface {
	Enqueue(item interface{})
	Dequeue() interface{}
	Size() int
	ReadAllMessages() []interface{}
	ClearQueue()
}