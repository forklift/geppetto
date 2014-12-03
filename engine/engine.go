package engine

import (
	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func New() (*Engine, error) {

	e := &Engine{
	//	make(chan unit.Event),
	}

	return e, nil
}

type Engine struct {
	//A units registery, holds all the units.
	units map[string]*Transaction

	//All le transactions.
	Transactions map[string]*Transaction
	Events       chan<- *event.Event

	//Internals.
	//One transaction at a time?
	//       lock sync.Mutex
}

func (e *Engine) Start(t *unit.Unit) error {

	//One Transaction at a time?
	//e.lock.Lock()
	//defer e.lock.Unlock()
	var transaction *Transaction

	if u, ok := e.units[t.Name]; ok {
		e.Events <- event.NewEvent(u.unit.Name, event.StatusAlreadyLoaded)

		e.Events <- event.NewEvent(u.unit.Name, event.StatusTransactionRegistering)
		//TODO: Health check? Status Check?
		e.Transactions[u.unit.Name] = u
		e.Events <- event.NewEvent(u.unit.Name, event.StatusTransactionRegistered)
		return nil

	} else {
		transaction = NewTransaction(e, t)
	}

	//TODO: prepare, health, add.
	return transaction.Start()

}

func (e *Engine) requestUnits(units []string) (map[string]*Transaction, error) {

	all := make(map[string]*Transaction)

	errs := make(chan error)
	cancel := make(chan struct{})
	transactions := prepareTransactions(e, errs, cancel, units)

	end := make(chan struct{})

	go func() {
		defer close(end)
		for t := range transactions {
			all[t.unit.Name] = t
		}
	}()

	select {
	case err := <-errs:
		if err != nil {
			close(cancel)

			//TODO: Clean up in the background. TODO: Pass the events? how do you log the progress?
			//go func() {
			<-end //Wait for all units.
			for _, t := range all {
				t.unit.Service.Cleanup() //TODO: Log/Handle errors
			}
			//	}()
			return nil, err //TODO: should retrun the units so fa?
		}
	case <-end:
		return all, nil
	}
	//TODO: Timeout?

	//TODO: We shouldn't really ever reach here. Panic? Error?
	return all, nil
}

func (e *Engine) requestUnit(name string) (*Transaction, error) {
	if u, ok := e.units[name]; ok {
		return u, nil
	}
	u, err := unit.Read(name)
	if err != nil {
		return nil, err
	}

	return NewTransaction(e, u), nil
}

func prepareTransactions(engine *Engine, errs chan<- error, cancel <-chan struct{}, names []string) <-chan *Transaction {

	transactions := make(chan *Transaction)
	go func() {
		defer close(transactions)
		for _, name := range names {

			transaction, err := engine.requestUnit(name)

			if err == nil {
				err := transaction.Prepare()
				if err != nil {
					errs <- err
					return
				}

				select {
				case transactions <- transaction:
					if err != nil {
						return
					}
				case <-cancel:
					return
				}

			}
		}
	}()

	return transactions
}
