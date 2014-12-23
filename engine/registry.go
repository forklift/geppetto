package engine

import (
	"sync"

	"github.com/forklift/operator/event"
	"github.com/forklift/operator/unit"
)

func NewRegistry() Registry {
	return Registry{units: make(map[string]*unit.Unit)}
}

type Registry struct {
	units map[string]*unit.Unit

	lock sync.Mutex
}

func (r *Registry) Request(from string, name string, out chan<- event.Event) (*unit.Unit, error) {
	r.lock.Lock()
	defer close(out)
	defer r.lock.Unlock()

	u, err := r.request(from, name, out)
	if err != nil {
		return nil, err
	}

	out <- event.New(u.Name, UnitRegistering, u.Name)
	u.Standalone = true

	return u, nil
}

func (r *Registry) request(from string, name string, out chan<- event.Event) (*unit.Unit, error) {

	out <- event.New(from, UnitLoading, name)

	u, ok := r.units[name]
	if ok {
		out <- event.New(from, UnitAlreadyLoaded, u.Name)
	} else {
		u := &unit.Unit{Name: name}
		err := unit.Read(u)
		if err != nil {
			out <- event.New(u.Name, UnitLoadingFailed, err.Error())
			return nil, err
		}

		r.units[name] = u
	}

	u.Deps = unit.NewUnitList()

	for _, dep := range append(u.Requires, u.Wants...) {
		d, err := r.request(u.Name, dep, out)
		if err != nil {
			return nil, err
		}

		u.Deps.Add(d)
	}

	return u, nil
}

func (r *Registry) Drop(name string, out chan<- event.Event) error {
	r.lock.Lock()
	defer close(out)
	defer r.lock.Unlock()

	delete(r.units, name)

	return nil
}

func (r *Registry) Load(out chan<- event.Event) error {
	r.lock.Lock()
	defer close(out)
	defer r.lock.Unlock()

	return nil
}
