package events

import "sync"

type EventName string
type EventListener func(_ ...interface{})

type Event struct {
	mu *sync.Mutex
	listeners map[EventName]EventListener
}

func Emitter() *Event {
	return &{
		listeners: make(map[EventName]EventListener)
	}
}

func (e *Event) Emit(eventName EventName, args ...interface{}) {
		e.mu.RLock()
		listener, ok := e.listeners[eventName]
		e.mu.RUnlock()
		if listener && ok {
			listener(args...)
		}
}

func (e *Event) On(eventName EventName,listener EventListener) {
		
		e.mu.Lock()
		defer e.mu.Unlock()
		e.listeners[eventName] = listener
}


func (e *Event) Off(eventName EventName) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.listeners, eventName)
}

func (ev *Event) Get(eventName EventName) EventListener {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.listeners[eventName]
}