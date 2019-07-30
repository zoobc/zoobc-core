package observer

import "sync"

type (
	Event string

	OnNotify func(data interface{}, args interface{})

	Listener struct {
		OnNotify OnNotify
	}

	Observer struct {
		Listeners map[Event][]Listener
	}
)

var (
	observerInstance  *Observer
	singletonObserver sync.Once
)

// NewObserver get single global instance of Observer
func NewObserver() *Observer {
	singletonObserver.Do(func() {
		observerInstance = &Observer{}
	})
	return observerInstance
}

// AddListener to add new listener in observer
func (o *Observer) AddListener(listener Listener, event Event) {
	if o.Listeners == nil {
		o.Listeners = map[Event][]Listener{}
	}
	o.Listeners[event] = append(o.Listeners[event], listener)
}

// Remove to remove registered listener in observer
func (o *Observer) Remove(event Event) {
	delete(o.Listeners, event)
}

// Notify to send data & arg to registered listener based on event
func (o *Observer) Notify(event Event, data, args interface{}) {
	for _, listener := range o.Listeners[event] {
		listener.OnNotify(data, args)
	}
}
