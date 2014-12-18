package unit

import "time"

type Check struct {
	Name     string
	Exec     string
	Interval time.Duration
}
