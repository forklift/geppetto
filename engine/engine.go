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

	out := make(chan event.Event)
	pipe := event.NewPipe()
	pipe.Add("Transaction", out)
	pipe.Add(e.name, e.Events)

	go func() {
		defer close(out)
		//Mark it as explicit.
		var u *unit.Unit

		pipe.Emit(event.New(e.name, event.UnitLoading, name))
		if u, ok := e.Units.Get(name); ok {
			pipe.Emit(event.New(e.name, event.UnitAlreadyLoaded, u.Name))
		} else {
			u, err := unit.New(name)

			if err != nil {
				//FIXME: too many parens. Define IO vs Parse error Events?
				pipe.Emit(event.New(u.Name, event.UnitLoadingFailed, fmt.Sprintf("Unit %s: %s", u.Name, err.Error())))
				return
			}
		}

		//TODO: Make this none blockiing and pipe the events out to pipe.
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

	return out
}

func (e *Engine) Kill(name string) chan event.Event {

	out := make(chan event.Event)
	pipe := event.NewPipe()
	pipe.Add("Transaction", out)
	pipe.Add(e.name, e.Events)

	go func() {
		defer close(out)

		u, ok := e.Units.Get(name)
		if !ok {
			pipe.Emit(event.New(e.name, event.UnitNotLoaded, u.Name))
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
		pipe.Pipe(u.Kill())
	}()

	return out
}
func (e *Engine) Prepare(u *unit.Unit) error {

	//e.Emit(event.New(u.Name, event.UnitPreparingFailed, u.Name, err))

	pipeline, errs, cancel, units := unit.NewPipeline()

	//Filter loaded units.
	loaded, fresh := pipeline.Filter(errs, cancel, units, e.Units.Has)

	//Load the fresh units.
	fresh = pipeline.Do(errs, cancel, fresh, unit.Read)

	//Attach the fresh ones to engine.
	fresh = pipeline.Do(errs, cancel, fresh, e.Attach)

	//Yup.
	all := pipeline.Merge(loaded, fresh)

	//Emit deps to units channel.
	//all = pipeline.Do(errs, cancel, all, pipeline.RequestDeps(units))

	//Prepare
	prepared := pipeline.Do(errs, cancel, all, pipeline.PrepareUnit)

	//Attach deps.
	withdeps := pipeline.Do(errs, cancel, prepared, pipeline.AttachDeps(prepared))

	// Start the pipeline...
	units <- u
	close(units) //TODO: What to do with the RequestDeps?

	return pipeline.Wait(errs, cancel, withdeps)
}
