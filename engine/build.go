package engine

import (
	"errors"
	"sync"

	"github.com/forklift/geppetto/unit"
)

func requestDeps(all <-chan *unit.Unit, units chan *unit.Unit) {
	var wg sync.WaitGroup
	go func() {

		defer close(units)
		//Pump all the dependencies.
		for u := range all {
			if u.Deps == nil { //If the Deps is set. we have already requested this units deps.
				go func() {
					wg.Add(1)
					u.Deps = unit.NewUnitList()
					for _, u := range unit.Make(append(u.Service.Requires, u.Service.Wants...)) {
						units <- u
					}
					wg.Done()
				}()
			}
		}
		wg.Wait()
	}()
}

func attachDeps(all <-chan *unit.Unit) func(*unit.Unit) error {
	return func(unit *unit.Unit) error {
		deps := append(unit.Service.Requires, unit.Service.Wants...)
		count := len(deps)
		for u := range all {
			for _, dep := range deps {
				if u.Name == dep {
					count--
					unit.Deps.Add(u)
					break
				}
			}

			if count == 0 {
				break
			}
		}

		if count != 0 {
			return errors.New("Missing dependency.")
		}
		return nil
	}
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
		//With "attachDeps" totally makes this worthy?
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

func wait(errs chan error, cancel chan struct{}, input <-chan *unit.Unit) error {

	//Wait for error or end.
	for {
		select {
		case err := <-errs:
			if err != nil {
				close(cancel)
				return err //TODO: multiple errors?
			}
		case <-input:
			if input == nil {
				return nil
			}
		}
	}
	//TODO: Timeout?
	//TODO: We shouldn't really ever reach here. Panic? Error?
	return nil
}
