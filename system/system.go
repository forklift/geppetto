package system

import (
	"github.com/forklift/geppetto/unit"
	"github.com/omeid/filebase"
	"github.com/omeid/filebase/codec"
)

const base = "/etc/geppetto"

func Units(query string) (map[string]*unit.Unit, error) {

	bucket := filebase.New(base, codec.INI{})

	objects, err := bucket.Objects(query, false)

	if err != nil {
		return nil, err
	}

	units := make(map[string]*unit.Unit)

	for _, name := range objects {
		u := &unit.Unit{}
		if err := bucket.Get(name, u); err != nil {
			return nil, err
		}

		units[name] = u
	}

	return units, nil
}

func Services(query string) (map[string]*unit.Unit, error) {
	return Units(query + ".service")
}
