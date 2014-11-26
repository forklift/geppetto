package events

import "github.com/forklift/geppetto/unit"

type Event struct {
	From    *unit.Unit
	Event   string
	Payload string
}
