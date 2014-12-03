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
}

func (e *Engine) Start(t *unit.Unit) error {

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
