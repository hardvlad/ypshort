// Package audit contains logic to send events to the external storage
package audit

import "sync"

type publisher interface {
	Register(observer)
	Deregister(observer)
	Notify()
}

type observer interface {
	Update(AuditorEvent)
	getID() string
}

// Event represents a publisher of events
type Event struct {
	observers map[string]observer
	data      AuditorEvent
	mu        sync.Mutex
}

// InitObserver creates a new Event object
func InitObserver() *Event {
	return &Event{}
}

// Register adds a new observer to the publisher
func (e *Event) Register(o observer) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.observers == nil {
		e.observers = make(map[string]observer)
	}
	e.observers[o.getID()] = o
}

// Deregister removes an observer from the publisher
func (e *Event) Deregister(o observer) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.observers, o.getID())
}

// Notify sends the event to all observers
func (e *Event) Notify() {
	for _, observer := range e.observers {
		observer.Update(e.data)
	}
}

// Update sends the event to all observers through a call to Notify
func (e *Event) Update(data AuditorEvent) {
	e.data = data
	e.Notify()
}
