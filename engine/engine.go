package engine

import (
	"errors"

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
	units map[string]*unit.Unit

	//All le transactions.
	Transactions map[string]*Transaction
	EventsPipe   chan unit.Event
}

type Status string

const (
	StatusLoaded        Status = "Loaded."
	StatusAlreadyLoaded Status = "Already Loaded."
)

func (e *Engine) HasUnit(name string) bool {
	_, ok := e.units[name]
	return ok
}

func (e *Engine) LoadUnit(u *unit.Unit) *unit.Unit {

	//TODO: Reload???
	if u, ok := e.units[u.Name]; !ok {
		e.units[u.Name] = u
	}

	return e.units[u.Name]

}

func (e *Engine) StartTransaction(u *unit.Unit) error {

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

func (e *Engine) StartUnit(name string) error {

	var (
		u  *unit.Unit
		ok bool
	)

	if u, ok = e.units[name]; !ok {
		return errors.New("Unloaded unit.")
	}

	u.Start()
	return nil
}
