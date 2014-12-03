package engine

import (
	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func NewTransaction(e *Engine, u *unit.Unit) *Transaction {
	return &Transaction{engine: e, unit: u}
}

type Transaction struct {
	engine *Engine
	unit   *unit.Unit
	ch     chan<- *event.Event

	//Internals
	prepared bool
}

func (t *Transaction) Prepare() error {

	if t.prepared {
		return nil
	}

	/*if t.engine.HasUnit(t.unit.Name) {
		t.engine.Events <- event.NewEvent(t.unit.Name, event.StatusAlreadyLoaded)
		return errors.New("Unit is already loaded. Transaction canceled.")
	}*/

	depslist := []string{}
	depslist = append(depslist, t.unit.Service.Requires...)
	depslist = append(depslist, t.unit.Service.Wants...)

	deps, err := t.engine.requestUnits(depslist)
	if err != nil {
		return err
	}
	_ = deps

	err = t.unit.Prepare()
	if err != nil {
		t.unit.Service.Cleanup()
		return err
	}

	//NOTE: Is this a good idea? Can we attempt to reprepare a transaction if it fails?
	if err == nil {
		t.prepared = true
	}
	return err
}

func (t *Transaction) Start() error {

	for _, name := range t.unit.Service.Before {
		_ = name
		//t.engine.Start(t.unit)
	}
	return nil
}
