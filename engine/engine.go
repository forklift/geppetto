package engine

import (
	"sync"

	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func New() (*Engine, error) {

	e := &Engine{
		Transactions: NewTransactionList(),
		Events:       make(chan *event.Event),
	}

	return e, nil
}

type Engine struct {

	//All le transactions.
	Transactions TransactionList

	Events chan *event.Event

	//Internals.
	//One transaction at a time
	lock sync.Mutex
}

func (e *Engine) Start(u *unit.Unit) error {

	//One Transaction at a time
	e.lock.Lock()
	defer e.lock.Unlock()
	var transaction *Transaction

	if t, ok := e.Transactions.Get(u.Name); ok {
		e.Events <- event.NewEvent(t.unit.Name, event.StatusAlreadyLoaded)

		e.Events <- event.NewEvent(t.unit.Name, event.StatusTransactionRegistering)
		//TODO: Health check? Status Check?
		t.Explicit = true
		e.Events <- event.NewEvent(t.unit.Name, event.StatusTransactionRegistered)
		return nil

	} else {
		transaction = NewTransaction(e, u)
		t.Explicit = true
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

	e.Transactions.Append(&transaction.deps)

	return nil
}
