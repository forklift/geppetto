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

func (ul *UnitList) addTo(list *UnitList) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	ul.ForEach(func(u *Unit) {
		list.units[u.Name] = u
	})
}

func (ul *UnitList) Append(list *UnitList) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	list.addTo(ul)
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

func (ul *UnitList) ForEach(do func(*Unit)) {
	ul.lock.Lock()
	defer ul.lock.Unlock()
	for _, u := range ul.units {
		do(u)
	}
}
