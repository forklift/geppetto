package unit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func New(name string) (*Unit, error) {
	u := &Unit{Name: name}
	return u, Read(u)
}

func Read(unit *Unit) error {
	location := filepath.Join(BasePath, unit.Name)
	unitfile, err := ioutil.ReadFile(location)
	if err != nil {
		return err
	}
	return Parse(unitfile, unit)
}

func Parse(unitfile []byte, unit *Unit) error {
	err := yaml.Unmarshal(unitfile, unit)
	fmt.Printf("err %+v\n", err)
	fmt.Printf("unit %+v\n", unit)
	os.Exit(0)

	return nil
}

func Make(names []string) []*Unit {
	units := []*Unit{}
	for _, name := range names {
		units = append(units, &Unit{Name: name})
	}
	return units
}
