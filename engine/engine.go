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

func (e *Engine) Attach(u *unit.Unit) error {
	e.Units.Add(u)
	return nil
}

func (e *Engine) Start(units ...*unit.Unit) chan *event.Event {

	out := make(chan *event.Event)
	pipe := event.NewPipe()
	pipe.Add("Transaction", out)
	pipe.Add("Geppetto", e.Events)

	go func() {
		pipe.Emit(event.New("Geppetto", event.StatusStartingService))
		for _, u := range units {

			if ut, ok := e.Units.Get(u.Name); ok {
				pipe.Emit(event.New(u.Name, event.StatusAlreadyLoaded))
				u = ut
			} else {
				err := e.Prepare(u)
				fmt.Printf("err %+v\n", err)
				if err != nil {
					pipe.Emit(event.New(u.Name, event.StatusFailed))
					break
				}

				u.Listeners.Add("Geppetto", e.Events)
				err = u.Start()
			}
			u.Explicit = true
			pipe.Emit(event.New(u.Name, event.StatusTransactionRegistering))
			u.Start()

		}
		close(out)
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
