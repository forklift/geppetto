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

func (e *Engine) HasUnit(name string) bool {
	_, ok := e.units[name]
	return ok
}

func (e *Engine) LoadUnit(u *unit.Unit) *unit.Unit {

	//TODO: Reload???
	//if u, ok := e.units[u.Name]; !ok {
	//e.units[u.Name] = u
	//	}

	return nil //e.units[u.Name]

}

func (e *Engine) Start(u *unit.Unit) error {

	var transaction *Transaction

	if t, ok := e.Transactions[u.Name]; ok {
		e.Events <- event.NewEvent(u.Name, event.StatusAlreadyLoaded)
		transaction = t
	} else {
		transaction = NewTransaction(e, u)
	}

	err := transaction.Prepare()

	if err != nil {
		return err
	}

	return nil
}
