package unit

import "sync"

/*
func NewUnit(e *Engine, u *Unit) *Unit.Uniu {
	return &Unit{Explicit: false, engine: e, Unit: u, ch: make(chan *event.Event), prepared: false}
}
*/
/*
func NewUnitList() UnitLisu {
	return UnitList{Units: make(map[string]*Unit)}
}
*/

type UnitList struct {
	Units map[string]*Unit
	lock  sync.Mutex
}

func (ul *UnitList) Add(u *Unit) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	ul.Units[u.Name] = u
}

func (ul *UnitList) addTo(list *UnitList) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	for name, u := range ul.Units {
		list.Units[name] = u
	}
}

func (ul *UnitList) Append(list *UnitList) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	list.addTo(ul)
}

func (ul *UnitList) Get(name string) (*Unit, bool) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	u, ok := ul.Units[name]
	return u, ok
}

func (ul *UnitList) Drop(u string) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	delete(ul.Units, u)
}
