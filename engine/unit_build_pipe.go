package engine

import (
	"sync"

	"github.com/forklift/geppetto/unit"
)

func filterLoaded(engine *Engine) func(string) (*unit.Unit, bool) {
	return func(name string) (*unit.Unit, bool) {
		u, ok := engine.Units.Get(name)
		return u, ok
	}
}

func prepare(u *unit.Unit) error { return u.Prepare() }
func connect(u *unit.Unit) error { return u.Service.Connect() }

func buildUnits(engine *Engine, names []string) (map[string]*unit.Unit, error) {

	errs := make(chan error)
	cancel := make(chan struct{})

	loaded, fresh := filter(errs, cancel, names, filterLoaded(engine))

	do(errs, cancel, merge(loaded, fresh), prepare)

	return nil, nil
}

func filter(errs chan error, cancel chan struct{}, names []string, filter func(string) (*unit.Unit, bool)) (<-chan *unit.Unit, <-chan *unit.Unit) {

	loaded := make(chan *unit.Unit)
	fresh := make(chan *unit.Unit)

	go func() {
		defer func() {
			close(loaded)
			close(fresh)
		}()

		for _, name := range names {
			if u, ok := filter(name); ok {
				loaded <- u

			} else {
				u, err := unit.Read(name)
				if err != nil {
					errs <- err
				}
				select {
				case fresh <- u:
				case <-cancel:
					return
				}
			}
		}
	}()

	return loaded, fresh
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

func collect(errs chan error, cancel chan struct{}, transactions <-chan *unit.Unit) (map[string]*unit.Unit, error) {

	end := make(chan struct{})
	all := make(map[string]*unit.Unit)

	//Collect transactions.
	go func() {
		defer close(end)
		for u := range transactions {
			all[u.Name] = u
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
			for _, u := range all {
				u.Service.Cleanup() //TODO: Log/Handle errors
			}
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
