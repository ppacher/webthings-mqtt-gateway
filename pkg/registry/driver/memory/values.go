package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/ppacher/mqtt-home/controller/pkg/mutex"
)

type itemValues struct {
	l *mutex.Mutex

	thingID string
	itemID  string

	values     []interface{}
	timestamps []time.Time
}

func newItemValues(thingID string, itemID string) *itemValues {
	return &itemValues{
		l:       mutex.New(),
		thingID: thingID,
		itemID:  itemID,
	}
}

func (iv *itemValues) Put(ctx context.Context, val interface{}) error {
	if !iv.l.TryLock(ctx) {
		return ctx.Err()
	}
	defer iv.l.Unlock()

	iv.values = append(iv.values, val)
	iv.timestamps = append(iv.timestamps, time.Now())

	return nil
}

func (iv *itemValues) Current(ctx context.Context) (interface{}, error) {
	if !iv.l.TryLock(ctx) {
		return nil, ctx.Err()
	}
	defer iv.l.Unlock()

	if len(iv.values) == 0 {
		return nil, nil
	}

	last := iv.values[len(iv.values)-1]

	return last, nil
}

func (iv *itemValues) Filter(ctx context.Context, from time.Time, to time.Time) (<-chan interface{}, error) {
	if !iv.l.TryLock(ctx) {
		return nil, ctx.Err()
	}
	defer iv.l.Unlock()

	return nil, fmt.Errorf("not yet implemented")
}

func (iv *itemValues) Clear(ctx context.Context) error {
	if !iv.l.TryLock(ctx) {
		return ctx.Err()
	}
	defer iv.l.Unlock()

	iv.values = nil
	iv.timestamps = nil

	return nil
}
