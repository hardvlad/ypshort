package audit

type publisher interface {
	Register(observer)
	Deregister(observer)
	Notify()
}

type observer interface {
	Update(AuditorEvent)
	getID() string
}

type Event struct {
	observers map[string]observer
	data      AuditorEvent
}

func InitObserver() *Event {
	return &Event{}
}

func (e *Event) Register(o observer) {
	if e.observers == nil {
		e.observers = make(map[string]observer)
	}
	e.observers[o.getID()] = o
}

func (e *Event) Deregister(o observer) {
	delete(e.observers, o.getID())
}

func (e *Event) Notify() {
	for _, observer := range e.observers {
		observer.Update(e.data)
	}
}

func (e *Event) Update(data AuditorEvent) {
	e.data = data
	e.Notify()
}
