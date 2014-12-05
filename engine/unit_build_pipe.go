package engine

import (
	"sync"

	"github.com/forklift/geppetto/unit"
)

func filterLoaded(engine *Engine) func(unit *unit.Unit) bool {
	return func(unit *unit.Unit) bool {
		_, ok := engine.Units.Get(unit.Name)
		return ok
	}
}

func prepare(engine *Engine) func(u *unit.Unit) error {

	return func(u *unit.Unit) error {
		err := u.Prepare()
		if err != nil {
			return err
		}

		return nil
	}
}

func connect(u *unit.Unit) error {
	return u.Service.Connect()
}

func start(engine *Engine, u *unit.Unit) (*unit.Unit, error) {

	//Mark it as user package.
	u.Explicit = true

	units := make(chan *unit.Unit)

	errs := make(chan error)
	cancel := make(chan struct{})

	loaded, fresh := filter(errs, cancel, units, filterLoaded(engine))

	all := merge(loaded, do(errs, cancel, fresh, unit.Read))

	//push the first unit.
	units <- u
	pushDeps(all, units)

	prepared := do(errs, cancel, all, prepare(engine))
	connected := do(errs, cancel, prepared, connect)

	deps, err := collect(errs, cancel, connected)
	if err != nil {
		return nil, err
	}

	//Do start.
	err = u.Start()

	if err != nil {
		return nil, err
	}

	//Add them to engine.
	engine.Units.Append(deps)

	return nil, nil //build(engine, errs, cancel, all)
}

func pushDeps(all <-chan *unit.Unit, units chan *unit.Unit) {
	var wg sync.WaitGroup
	go func() {

		defer close(units)
		//Pump all the dependencies.
		for u := range all {
			wg.Add(1)
			go func() {
				for _, u := range unit.Make(append(u.Service.Requires, u.Service.Wants...)) {
					units <- u
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}()
}

// MapReduce-ish helpers.
func filter(errs chan error, cancel chan struct{}, units <-chan *unit.Unit, filter func(*unit.Unit) bool) (<-chan *unit.Unit, <-chan *unit.Unit) {

	pass := make(chan *unit.Unit)
	fail := make(chan *unit.Unit)

	go func() {
		defer func() {
			close(pass)
			close(fail)
		}()

		for u := range units {
			if filter(u) {
				pass <- u

			} else {
				select {
				case fail <- u:
				case <-cancel:
					return
				}
			}
		}
	}()

	return pass, fail
}

func do(errs chan error, cancel chan struct{}, units <-chan *unit.Unit, do func(*unit.Unit) error) <-chan *unit.Unit {

	prepared := make(chan *unit.Unit)

	go func() {
		defer close(prepared)
		//TODO: We can forkout a new goroutines for every "task", but is it worth it? (similar to "merge").
		for unit := range units {

			err := do(unit)
			if err != nil {
				errs <- err
				return
			}

			select {
			case prepared <- unit:
			case <-cancel:
				return
			}
		}
	}()

	return prepared
}

func merge(transactionChans ...<-chan *unit.Unit) <-chan *unit.Unit {

	transactions := make(chan *unit.Unit)

	var wg sync.WaitGroup
	wg.Add(len(transactionChans))

	for _, ch := range transactionChans {
		go func(uc <-chan *unit.Unit) {
			for t := range ch {
				transactions <- t
			}
			wg.Done()
		}(ch)

	}

	go func() {
		wg.Wait()
		close(transactions)
	}()

	return transactions
}

func collect(errs chan error, cancel chan struct{}, transactions <-chan *unit.Unit) (*unit.UnitList, error) {

	end := make(chan struct{})
	all := unit.NewUnitList()

	//Collect transactions.
	go func() {
		defer close(end)
		for u := range transactions {
			all.Add(u)
		}
		close(end) //End.
	}()

	//Wait for error or end.
	select {
	case err := <-errs:
		if err != nil {
			close(cancel)
			//TODO: Clean up in the background. TODO: Pass the events? how do you log the progress?
			//go func() {
			<-end //Wait for all units.
			all.ForEach(func(u *unit.Unit) { u.Service.Cleanup() })

			// }()
			return nil, err //TODO: should retrun the units so fa?
		}
	case <-end:
		return all, nil
	}
	//TODO: Timeout?
	//TODO: We shouldn't really ever reach here. Panic? Error?
	return all, nil
}
