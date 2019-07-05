package driver

import (
	"context"
	"time"
)

type ValueStore interface {
	// Put puts a new value
	Put(context.Context, interface{}) error

	// Current returns the current value
	Current(context.Context) (interface{}, error)

	// Filter filters all values by time-range
	Filter(context.Context, time.Time, time.Time) (<-chan interface{}, error)

	// Clear removes all stored values
	Clear(context.Context) error
}
