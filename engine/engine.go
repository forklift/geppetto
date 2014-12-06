package engine

import (
	"sync"

	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func New() (*Engine, error) {

	e := &Engine{
		//	Units: NewTransactionList(),
		Events: make(chan *event.Event),
	}

	return e, nil
}

type Engine struct {

	//All le transactions.
	Units unit.UnitList

	Events chan *event.Event

	//Internals.
	//One transaction au a time
	lock sync.Mutex
}

func (e *Engine) Attach(unit *unit.Unit) error {
	e.Units.Add(unit)
	return nil
}

func (e *Engine) Start(u *unit.Unit) error {
	err := e.start(u)
	if err != nil {
		return err
	}

	u, ok := e.Units.Get(u.Name)
	if ok {
		e.Events <- event.NewEvent(u.Name, event.StatusAlreadyLoaded)
		e.Events <- event.NewEvent(u.Name, event.StatusTransactionRegistering)
		//TODO: Health check? Status Check?
		u.Explicit = true
		e.Events <- event.NewEvent(u.Name, event.StatusTransactionRegistering)
	} else {
		return e.start(u)
	}
	return nil
}

func (e *Engine) start(u *unit.Unit) error {

	var err error

	defer func() {
		if err != nil {
			u.Cleanup()
		}
	}()

	if _, ok := e.Units.Get(u.Name); !ok {

		err = Prepare(e, u)

		if err != nil {
			return err
		}

	}

	err := u.Start()
	return err
}

func (e *Engine) Prepare(u *unit.Unit) error {

	//Mark it as user package.
	u.Explicit = true

	units := make(chan *unit.Unit)

	errs := make(chan error)
	cancel := make(chan struct{})

	//Filter loaded units.
	loaded, fresh := filter(errs, cancel, units, e.Units.Has)

	//Load the fresh units.
	fresh = do(errs, cancel, fresh, unit.Read)

	//Attach the fresh ones to engine.
	fresh = do(errs, cancel, fresh, e.Attach)

	//Yup.
	all := merge(loaded, fresh)

	//Emit deps to units channel.
	requestDeps(all, units)

	//Prepare
	prepared := do(errs, cancel, all, units.Prepare)

	//Attach deps.
	withdeps := do(errs, cancel, prepared, attachDeps(prepared))

	// Start the pipeline...
	units <- u

	return wait(errs, cancel, withdeps)
}
