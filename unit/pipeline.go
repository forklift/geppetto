package unit

import "sync"

func NewPipeline() (Pipeline, chan error, chan struct{}, chan *Unit) {
	return Pipeline{}, make(chan error), make(chan struct{}), make(chan *Unit)
}

//Filter loaded units.
//	loaded, fresh := filter(errs, cancel, units, e.Units.Has)
// MapReduce-ish helpers.

//It is merely used as a "namespace" as we can't put it as a package since it will result in cyclic import of pipeline/unit as we need pipeline for Unit.Start.
type Pipeline struct{}

func (p Pipeline) Filter(errs chan error, cancel chan struct{}, units <-chan *Unit, filter func(*Unit) bool) (<-chan *Unit, <-chan *Unit) {

	pass := make(chan *Unit)
	fail := make(chan *Unit)

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

func (p Pipeline) Do(errs chan error, cancel chan struct{}, units <-chan *Unit, do func(*Unit) error) <-chan *Unit {

	prepared := make(chan *Unit)

	go func() {
		defer close(prepared)
		var wg sync.WaitGroup

		for unit := range units {
			wg.Add(1)
			go func() {
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
			}()
		}
		wg.Wait()

	}()

	return prepared
}

func (p Pipeline) Merge(transactionChans ...<-chan *Unit) <-chan *Unit {

	transactions := make(chan *Unit)

	var wg sync.WaitGroup
	wg.Add(len(transactionChans))

	for _, ch := range transactionChans {
		go func(uc <-chan *Unit) {
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

func (p Pipeline) Wait(errs chan error, cancel chan struct{}, input <-chan *Unit) error {

	//Wait for error or end.
	//TODO: Timeout?
	for {
		select {
		case err := <-errs:
			if err != nil {
				close(cancel)
				return err //TODO: multiple errors?
			}
		case _, open := <-input:
			if !open {
				return nil
			}
		}
	}
	//TODO: We shouldn't really ever reach here. Panic? Error?
	return nil
}

//Helpers

func (p Pipeline) PrepareUnit(u *Unit) error {

	err := u.Prepare()
	if err != nil {
		return err
	}
	return nil
}

func (p Pipeline) StartUnit(u *Unit) error {
	u.Start()
	return nil
}

func (p Pipeline) RequestDeps(units chan *Unit) func(*Unit) error {
	return func(u *Unit) error {
		u.Deps = NewUnitList()
		for _, u := range Make(append(u.Service.Requires, u.Service.Wants...)) {
			units <- u
		}
		return nil
	}
}

func (p Pipeline) AttachDeps(all <-chan *Unit) func(*Unit) error {
	return func(unit *Unit) error {
		/*
			deps := append(unit.Service.Requires, unit.Service.Wants...)
			count := len(deps)
			for u := range all {
				for _, dep := range deps {
					if u.Name == dep {
						count--
						unit.Deps.Add(u)
						u.Listeners.Add(unit.Name, unit.Ch)
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
		*/
		return nil
	}
}
