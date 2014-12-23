package engine

import "github.com/forklift/operator/event"

func (e *Engine) Start(name string) (<-chan event.Event, chan error) {
	return e.job(e.start(name))
}

func (e *Engine) start(name string) func(out chan<- event.Event) error {
	return func(out chan<- event.Event) error {
		defer close(out)

		return nil
	}
}
