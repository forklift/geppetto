package engine

import (
	"sync"

	"github.com/forklift/operator/event"
)

func New() *Engine {

	e := &Engine{
		name:     "Geppetto",
		Registry: NewRegistry(),
		Events:   make(chan event.Event),
		//Listeners: event.NewPipe(),
	}

	//go e.Listeners.Watch(e.Events)

	return e
}

type Engine struct {
	name string

	Registry Registry
	Events   chan event.Event
	//Internals.
	//One transaction au a time
	lock sync.Mutex
}

const (
	UnitLoading       event.Type = "Loading Unit."
	UnitLoadingFailed event.Type = "Unit Load failed."
	UnitRegistering   event.Type = "Registering Unit."
	UnitNotLoaded     event.Type = "Unit Not Loaded."
	UnitAlreadyLoaded event.Type = "Unit Already loaded."
)

func (e *Engine) job(job func(chan<- event.Event) error) (<-chan event.Event, chan error) {
	out := make(chan event.Event)
	errch := make(chan error)

	go func() {
		defer close(errch)
		err := job(out)
		if err != nil {
			errch <- err
		}
	}()

	return out, errch
	//return event.Pipe(transaction, e.Events), errch ??
}
