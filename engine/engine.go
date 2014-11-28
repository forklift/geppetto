package engine

import "github.com/forklift/geppetto/unit"

func New() {}

func Start(u unit.Unit) error {

	err := u.Prepare()
	return err
}
