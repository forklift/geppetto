package engine

import (
	"sync"

	"github.com/forklift/geppetto/event"
	"github.com/forklift/geppetto/unit"
)

func NewTransaction(e *Engine, u *unit.Unit) *Transaction {
	return &Transaction{Explicit: false, engine: e, unit: u, ch: make(chan *event.Event), prepared: false}
}

func NewTransactionList() TransactionList {
	return TransactionList{transactions: make(map[string]*Transaction)}
}

type TransactionList struct {
	transactions map[string]*Transaction
	lock         sync.Mutex
}

func (ts *TransactionList) Add(t *Transaction) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	ts.transactions[t.unit.Name] = t
}

func (ts *TransactionList) addTo(list *TransactionList) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	for name, t := range ts.transactions {
		list.transactions[name] = t
	}
}

func (ts *TransactionList) Append(list *TransactionList) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	list.addTo(ts)
}

func (ts *TransactionList) Get(t string) (*Transaction, bool) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	u, ok := ts.transactions[t]
	return u, ok
}

func (ts *TransactionList) Drop(t string) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	delete(ts.transactions, t)
}

type Transaction struct {
	//An Explicit transaction is directly started by the Engine and is not a dep.
	// So it should not be "terminated" as dependency.
	Explicit bool

	engine *Engine
	unit   *unit.Unit
	ch     chan *event.Event

	depslist []string

	deps     TransactionList
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

	t.depslist = append(t.unit.Service.Requires, t.unit.Service.Wants...)

	for _, name := range t.depslist {
		t, ok := t.engine.Transactions.Get(name)
		if ok {
			t.deps.Add(t)
			continue
		}
		//Prepare a new transaction of the dep.
	}

	err := t.unit.Prepare()
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
