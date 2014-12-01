package engine

import (
	"sync"

	"github.com/forklift/geppetto/unit"
)

func NewTransaction(e *Engine, u *unit.Unit) *Transaction {
	return &Transaction{engine: e, unit: u, deps: make(map[string]*unit.Unit)}
}

type Transaction struct {
	engine *Engine
	unit   *unit.Unit
	deps   map[string]*unit.Unit
}

func (t *Transaction) Prepare() error {

	var err error
	t.deps, err = buildUnits(t.engine, t.unit.Service.Requires, t.unit.Service.Wants)
	if err != nil {
		return err
	}

	err = t.unit.Prepare()
	if err != nil {
		return err
	}

	return err
}

func (t *Transaction) Start() error {

	for _, name := range t.unit.Service.Before {
		t.engine.StartUnit(name)
	}
	return nil
}

func readUnits(engine *Engine, errs chan error, names []string) chan *unit.Unit {
	units := make(chan *unit.Unit)

	go func() {
		defer close(units)

		for _, name := range names {

			var (
				u   *unit.Unit
				err error
			)
			if !engine.HasUnit(name) {
				u, err = unit.Read(name)
				if err != nil {
					errs <- err
					close(errs)
				}
			}

			select {
			case units <- u:
			case e := <-errs:
				errs <- e
				return
			}
		}
	}()
	return units

}

func prepareUnits(errs chan error, units chan *unit.Unit) chan *unit.Unit {

	prepared := make(chan *unit.Unit)

	defer close(units)

	go func() {
		for unit := range units {
			err := unit.Prepare()
			if err != nil {
				unit.Service.Cleanup()
				close(errs)
			}
			return
			select {
			case units <- unit:
			case e := <-errs:
				errs <- e
				return
			}

		}
	}()

	return prepared
}

func mergeUnits(errs chan error, ucs ...chan *unit.Unit) chan *unit.Unit {

	all := make(chan *unit.Unit)

	var wg sync.WaitGroup
	wg.Add(len(ucs))

	for _, uc := range ucs {
		go func(uc chan *unit.Unit) {
			for {
				select {
				case u := <-uc:
					all <- u
				case e := <-errs:
					errs <- e
					break
				}
			}
			wg.Done()
		}(uc)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(errs)
	}()

	return all
}

func buildUnits(engine *Engine, unitlists ...[]string) (map[string]*unit.Unit, error) {

	errs := make(chan error)

	all := make(map[string]*unit.Unit)

	prepared := make([]chan *unit.Unit, len(unitlists))

	for _, units := range unitlists {
		prepared = append(prepared, prepareUnits(errs, readUnits(engine, errs, units)))
	}

	for {
		select {
		case u := <-mergeUnits(errs, prepared...):
			all[u.Name] = u
		case err := <-errs:
			if err == nil {
				return all, nil
			}
			for _, u := range all {
				u.Service.Cleanup() //TODO: Log/Handle errors
			}
			return nil, err //TODO: should retrun the units so fa?
		}
	}
	//TODO: We shouldn't really ever reach here. Panic? Error?
	return all, nil
}
