package clients

import (
	"sync"
)

type ClientEvent struct {
	ClientID uint32
}

type ClientEventHandler func(event ClientEvent)

type ClientEventManager struct {
	lock     sync.Mutex
	handlers []ClientEventHandler
}

func NewClientEventManager() *ClientEventManager {
	return &ClientEventManager{}
}

// RegisterHandler registers a handler for events.
// The handler will be called in a goroutine.
func (em *ClientEventManager) RegisterHandler(handler ClientEventHandler) {
	em.lock.Lock()
	defer em.lock.Unlock()
	em.handlers = append(em.handlers, handler)
}

// Trigger triggers an event.
// All registered handlers will be called their own goroutine.
func (em *ClientEventManager) Trigger(event ClientEvent) {
	em.lock.Lock()
	defer em.lock.Unlock()
	for _, handler := range em.handlers {
		go handler(event)
	}
}
