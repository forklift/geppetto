package engine

import (
	"fmt"
	"sync"

	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func New() *Engine {

	e := &Engine{
		Units:  unit.NewUnitList(),
		Events: make(chan *event.Event),
	}

	return e
}

type Engine struct {

	//All le transactions.
	Units *unit.UnitList

	Events chan *event.Event

	//Internals.
	//One transaction au a time
	lock sync.Mutex
}

func (e *Engine) Attach(unit *unit.Unit) error {
	e.Units.Add(unit)
	return nil
}

func (e *Engine) Start(units ...*unit.Unit) chan *event.Event {

	fmt.Println("Starting...")
	out := make(chan *event.Event)
	pipe := event.NewPipe()
	pipe.Add("Transaction", out)
	//pipe.Add("Geppetto", e.Events)

	go func() {
		fmt.Println("Going bg...")
		pipe.Emit(event.NewEvent("Geppetto", event.StatusStartingService))
		for _, u := range units {

			if ut, ok := e.Units.Get(u.Name); ok {
				pipe.Emit(event.NewEvent(u.Name, event.StatusAlreadyLoaded))
				u = ut
			} else {

				fmt.Println("Prepareing...")
				err := e.Prepare(u)
				if err != nil {
					pipe.Emit(event.NewEvent(u.Name, event.StatusFailed))
				}

				err = u.Start()
			}
			u.Listeners.Add("Geppetto", e.Events)
			u.Explicit = true
			pipe.Emit(event.NewEvent(u.Name, event.StatusTransactionRegistering))

		}
	}()

	return out
}

func (e *Engine) Prepare(u *unit.Unit) error {

	//Mark it as user package.
	u.Explicit = true

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
	pipeline.RequestDeps(all, units)

	//Prepare
	prepared := pipeline.Do(errs, cancel, all, pipeline.PrepareUnit)

	//Attach deps.
	withdeps := pipeline.Do(errs, cancel, prepared, pipeline.AttachDeps(prepared))

	// Start the pipeline...
	units <- u

	return pipeline.Wait(errs, cancel, withdeps)
}
