package engine

import "github.com/forklift/geppetto/unit"

type Transaction struct {
	unit *unit.Unit
	deps map[string]*unit.Unit
}
