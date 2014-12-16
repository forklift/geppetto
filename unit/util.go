package unit

import (
	"io"
	"os"
	"path/filepath"

	"github.com/omeid/go-ini"
)

func New(name string) (*Unit, error) {
	u := &Unit{Name: name}
	return u, Read(u)
}

func Read(unit *Unit) error {

	file, err := os.Open(filepath.Join(BasePath, unit.Name))
	if err != nil {
		return err
	}

	return Parse(file, unit)
}

func Parse(reader io.Reader, unit *Unit) error {
	return ini.NewDecoder(reader).Decode(unit)
}

func Make(names []string) []*Unit {
	units := []*Unit{}
	for _, name := range names {
		units = append(units, &Unit{Name: name})
	}
	return units
}
