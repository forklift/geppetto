package engine

import (
	"fmt"
	"sync"

	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func New() *Engine {

	e := &Engine{
		name:   "Geppetto",
		Units:  unit.NewUnitList(),
		Events: make(chan event.Event),
	}

	return e
}

type Engine struct {
	name string
	//All le transactions.
	Units *unit.UnitList

	Events chan event.Event

	//Internals.
	//One transaction au a time
	lock sync.Mutex
}

func (e *Engine) Attach(u *unit.Unit) error {
	e.Units.Add(u)
	return nil
}

func (e *Engine) Start(name string) chan event.Event {

	transaction := make(chan event.Event)
	pipe := event.NewPipe()
	pipe.Add("Transaction", transaction)
	pipe.Add(e.name, e.Events)

	go func() {
		defer close(transaction)
		//Mark it as explicit.
		u := &unit.Unit{Name: name}

		pipe.Emit(event.New(e.name, event.UnitLoading, name))
		if ut, ok := e.Units.Get(u.Name); ok {
			pipe.Emit(event.New(e.name, event.UnitAlreadyLoaded, u.Name))
			u = ut
		} else {
			err := unit.Read(u)
			if err != nil {
				//FIXME: too many parens. Define IO vs Parse error Events?
				pipe.Emit(event.New(u.Name, event.UnitLoadingFailed, fmt.Sprintf("Unit %s: %s", u.Name, err.Error())))
				return
			}
		}

		//TODO: Make this none blockiing and pipe the events out to pipe similar to u.Start
		err := e.Prepare(u)
		if err != nil {
			pipe.Emit(event.New(u.Name, event.UnitPreparingFailed, err.Error()))
			return
		}

		if !u.Explicit {
			u.Explicit = true
			pipe.Emit(event.New(e.name, event.UnitRegistering, u.Name))
		}

		//This is blocking.
		pipe.Pipe(u.Start())

		//Add Engine to the unit listeners.
		u.Listeners.Add(e.name, u.Events)
	}()

	return transaction
}

func (e *Engine) Prepare(u *unit.Unit) error {

	transaction := make(chan event.Event)
	pipe := event.NewPipe()
	pipe.Add("Transaction", transaction)
	pipe.Add(e.name, e.Events)
	//e.Emit(event.New(u.Name, event.UnitPreparingFailed, u.Name, err))]
	pipeline, errs, cancel, units := unit.NewPipeline()

	var count sync.WaitGroup
	all := pipeline.Do(errs, cancel, units, func(*unit.Unit) error {
		count.Add(1)
		return nil
	})
	//Filter loaded units.
	loaded, fresh := pipeline.Filter(errs, cancel, all, e.Units.Has)

	//Load the fresh units.
	fresh = pipeline.Do(errs, cancel, fresh, unit.Read)

	//Attach the fresh ones to engine.
	fresh = pipeline.Do(errs, cancel, fresh, e.Attach)

	//Yup.
	all = pipeline.Merge(loaded, fresh)

	//Emit deps to units channel.
	all = pipeline.Do(errs, cancel, all, pipeline.RequestDeps(units))

	//Once a unit is here, it has it's Deps already sent out.
	//Drop the count.
	all = pipeline.Do(errs, cancel, all, func(*unit.Unit) error {
		count.Done()
		return nil
	})

	//Prepare
	prepared := pipeline.Do(errs, cancel, all, pipeline.PrepareUnit)

	//Attach deps.
	//Pass the wait group.

	// Start the pipeline...
	units <- u

	//Wait for all units to have their Deps proccessed.
	count.Wait()

	//Close the pipeline.
	close(units)

	//Wait for the pipeline to finish.
	return pipeline.Wait(errs, cancel, prepared)
}

func (e *Engine) Stop(name string) chan event.Event {

	out := make(chan event.Event)
	pipe := event.NewPipe()
	pipe.Add("Transaction", out)
	pipe.Add(e.name, e.Events)

	go func() {
		defer close(out)

		u, ok := e.Units.Get(name)
		if !ok {
			pipe.Emit(event.New(e.name, event.UnitNotLoaded, name))
			return
		}

		if !u.Explicit {
			out <- event.New(e.name, event.ForbiddenOperation, "Unit is a child. Won't compley. Talk to parents.")
			return
		}

		u.Explicit = false
		pipe.Emit(event.New(e.name, event.UnitDeregistered, u.Name))

		if u.Listeners.Count() > 1 {
			out <- event.New(e.name, event.ForbiddenOperation, "Unit is a dependency now.")
			return
		}

		pipe.Pipe(u.Stop(e.name))
		u.Deps.ForEach(func(u *unit.Unit) {
			if u.Listeners.Count() < 2 {
				e.Units.Drop(name)
			}
		})
		e.Units.Drop(name)

		//TODO: Clean up dropping
	}()

	return out
}
