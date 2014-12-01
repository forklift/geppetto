package engine

import "github.com/forklift/geppetto/unit"

func New() (*Engine, error) {

	e := &Engine{
	//	make(chan unit.Event),
	}

	return e, nil
}

type Engine struct {
	//A units registery, holds all the units.
	Units map[string]*unit.Unit

	//All le transactions.
	Transactions map[string]*Transaction
	EventsPipe   chan unit.Event
}

type Status string

const (
	StatusLoaded        Status = "Loaded."
	StatusAlreadyLoaded Status = "Already Loaded."
)

func (e *Engine) LoadUnit(u *unit.Unit) *unit.Unit {

	//TODO: Reload???
	if u, ok := e.Units[u.Name]; !ok {
		e.Units[u.Name] = u
	}

	return e.Units[u.Name]

}

func (e *Engine) Start(u *unit.Unit) error {

	var err error

	if _, ok := e.Transactions[u.Name]; !ok {
		t := NewTransaction(e, u)

		err = t.Prepare()
		//e.Transactions[u.Name]
	}

	if err != nil {
		return err
	}

	return nil
}
