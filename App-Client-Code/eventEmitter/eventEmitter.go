package eventEmitter

import (
	"reflect"
	"sync"
)

// Events represents a storage for event listeners.
type Events struct {
	listeners map[string][]*EE
	mutex     sync.RWMutex
}

// EE represents a single event listener.
type EE struct {
	fn      func(...interface{})
	context interface{}
	once    bool
}

// EventEmitter represents an event emitter.
type EventEmitter struct {
	events Events
}

// NewEventEmitter creates a new EventEmitter.
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{events: Events{listeners: make(map[string][]*EE)}}
}

// On adds a listener for a given event.
func (emitter *EventEmitter) On(event string, fn func(...interface{}), context interface{}) {
	emitter.events.mutex.Lock()
	defer emitter.events.mutex.Unlock()

	emitter.events.listeners[event] = append(emitter.events.listeners[event], &EE{fn: fn, context: context, once: false})
}

// Once adds a one-time listener for a given event.
func (emitter *EventEmitter) Once(event string, fn func(...interface{}), context interface{}) {
	emitter.events.mutex.Lock()
	defer emitter.events.mutex.Unlock()

	emitter.events.listeners[event] = append(emitter.events.listeners[event], &EE{fn: fn, context: context, once: true})
}

// Emit calls each of the listeners registered for a given event.
func (emitter *EventEmitter) Emit(event string, args ...interface{}) bool {
	emitter.events.mutex.RLock()
	defer emitter.events.mutex.RUnlock()

	listeners, ok := emitter.events.listeners[event]
	if !ok {
		return false
	}

	for _, listener := range listeners {
		if listener.once {
			emitter.RemoveListener(event, listener.fn)
		}

		listener.fn(args...)
	}

	return true
}

// RemoveListener removes the listeners of a given event.
func (emitter *EventEmitter) RemoveListener(event string, fn func(...interface{})) {
	emitter.events.mutex.Lock()
	defer emitter.events.mutex.Unlock()

	listeners, ok := emitter.events.listeners[event]
	if !ok {
		return
	}

	var remainingListeners []*EE
	for _, listener := range listeners {
		if listener.fn != nil && reflect.ValueOf(listener.fn).Pointer() != reflect.ValueOf(fn).Pointer() {
			remainingListeners = append(remainingListeners, listener)
		}
	}

	emitter.events.listeners[event] = remainingListeners
}

// RemoveAllListeners removes all listeners, or those of the specified event.
func (emitter *EventEmitter) RemoveAllListeners(event string) {
	emitter.events.mutex.Lock()
	defer emitter.events.mutex.Unlock()

	if event != "" {
		delete(emitter.events.listeners, event)
	} else {
		emitter.events.listeners = make(map[string][]*EE)
	}
}
