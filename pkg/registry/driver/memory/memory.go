package memory

import (
	"context"
	"fmt"

	"github.com/ppacher/mqtt-home/controller/pkg/mutex"
	"github.com/ppacher/mqtt-home/controller/pkg/registry/driver"
	"github.com/ppacher/mqtt-home/controller/pkg/spec"
)

func init() {
	driver.MustRegister("memory", func(_ string) (driver.Driver, error) {
		return New(), nil
	})
}

type memDriver struct {
	m          *mutex.Mutex
	things     map[string]*spec.Thing
	itemStores map[string]*itemValues
}

// New returns a new memory driver
func New() driver.Driver {
	return &memDriver{
		m:          mutex.New(),
		things:     make(map[string]*spec.Thing),
		itemStores: make(map[string]*itemValues),
	}
}

func (mem *memDriver) Get(ctx context.Context, id string) (*spec.Thing, error) {
	if !mem.m.TryLock(ctx) {
		return nil, ctx.Err()
	}
	defer mem.m.Unlock()

	thing, ok := mem.things[id]
	if !ok {
		return nil, driver.ErrUnknownThing
	}

	return thing, nil
}

func (mem *memDriver) Set(ctx context.Context, thing *spec.Thing, opts *driver.SetOptions) error {
	if opts == nil {
		opts = &driver.SetOptions{}
	}

	if opts.UpdateOnly && opts.CreateOnly {
		return driver.ErrInvalidOptions
	}

	if !mem.m.TryLock(ctx) {
		return ctx.Err()
	}
	defer mem.m.Unlock()

	if opts.CreateOnly || opts.UpdateOnly {
		_, ok := mem.things[thing.ID]

		if opts.CreateOnly && ok {
			return driver.ErrThingExists
		}

		if opts.UpdateOnly && !ok {
			return driver.ErrUnknownThing
		}
	}

	mem.things[thing.ID] = thing

	return nil
}

func (mem *memDriver) Delete(ctx context.Context, id string, opts *driver.DeleteOptions) (*spec.Thing, error) {
	if !mem.m.TryLock(ctx) {
		return nil, ctx.Err()
	}
	defer mem.m.Unlock()

	t, ok := mem.things[id]
	if opts != nil && opts.MustExist && !ok {
		return nil, driver.ErrUnknownThing
	}

	delete(mem.things, id)

	return t, nil
}

func (mem *memDriver) Has(ctx context.Context, id string) (bool, error) {
	// inside the memory driver a simple Has() check is not better than using Get()
	_, err := mem.Get(ctx, id)
	return err != driver.ErrUnknownThing, err
}

func (mem *memDriver) IDs(ctx context.Context) ([]string, error) {
	ids := []string{}

	if !mem.m.TryLock(ctx) {
		return nil, ctx.Err()
	}
	defer mem.m.Unlock()

	for id := range mem.things {
		ids = append(ids, id)
	}

	return ids, nil
}

func (mem *memDriver) ItemValues(ctx context.Context, thingID string, itemID string) (driver.ValueStore, error) {
	id := fmt.Sprintf("%s/%s", thingID, itemID)
	if !mem.m.TryLock(ctx) {
		return nil, ctx.Err()
	}
	defer mem.m.Unlock()

	store, ok := mem.itemStores[id]
	if !ok {
		store = newItemValues(thingID, itemID)
		mem.itemStores[id] = store
	}

	return store, nil
}
