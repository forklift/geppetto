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

func (e *Engine) Start(u *unit.Unit) error {

	if _, ok := e.Units[u.Name]; !ok {
		e.Units[u.Name] = u
	}

	err := u.Prepare()

	if err != nil {
		return err
	}

	return nil
}
