package unit

import "sync"

func NewUnitList() *UnitList {
	return &UnitList{units: make(map[string]*Unit)}
}

type UnitList struct {
	units map[string]*Unit
	lock  sync.Mutex
}

func (ul *UnitList) Add(u *Unit) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	ul.units[u.Name] = u
}

func (ul *UnitList) Merge(list *UnitList) {
	ul.lock.Lock()
	list.lock.Lock()

	defer ul.lock.Unlock()
	defer list.lock.Unlock()

	list.foreach(func(u *Unit) {
		ul.units[u.Name] = u
	})
}

func (ul *UnitList) Get(name string) (*Unit, bool) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	u, ok := ul.units[name]
	return u, ok
}

func (ul *UnitList) Drop(u string) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	delete(ul.units, u)
}

func (ul *UnitList) Has(unit *Unit) bool {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	_, ok := ul.units[unit.Name]
	return ok
}

func (ul *UnitList) ForEach(do func(*Unit)) *UnitList {
	ul.lock.Lock()
	defer ul.lock.Unlock()
	ul.foreach(do)
	return ul
}

func (ul *UnitList) Then(do func()) {
	do()
}

func (ul *UnitList) foreach(do func(*Unit)) {
	for _, u := range ul.units {
		do(u)
	}
}
