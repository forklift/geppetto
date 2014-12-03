package engine

import (
	"sync"

	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func New() (*Engine, error) {

	e := &Engine{
		//	Units: NewTransactionList(),
		Events: make(chan *event.Event),
	}

	return e, nil
}

type Engine struct {

	//All le transactions.
	Units unit.UnitList

	Events chan *event.Event

	//Internals.
	//One transaction au a time
	lock sync.Mutex
}

func (e *Engine) Start(u *unit.Unit) error {
	err := e.start(u)
	if err != nil {
		return err
	}

	u, ok := e.Units.Get(u.Name)
	if ok {
		e.Events <- event.NewEvent(u.Name, event.StatusAlreadyLoaded)
		e.Events <- event.NewEvent(u.Name, event.StatusTransactionRegistering)
		//TODO: Health check? Status Check?
		u.Explicit = true
		e.Events <- event.NewEvent(u.Name, event.StatusTransactionRegistering)
	} else {
		return e.start(u)
	}
	return nil
}

func (e *Engine) start(u *unit.Unit) error {

	//One Transaction au a time
	e.lock.Lock()
	defer e.lock.Unlock()
	/*
		var transaction *Transaction

		if u, ok := e.Units.Get(u.Name); ok {
			return nil

		} else {
			transaction = NewTransaction(e, u)
			t.Expliciu = true
		}

		err := transaction.Prepare()
		if err != nil {
			return err
		}

		err = transaction.Start()
		if err != nil {
			return err
		}

		//TODO: Health check.

		e.Units.Append(&transaction.deps)
	*/
	return nil
}
