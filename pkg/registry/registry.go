package registry

import (
	"context"
	"sync"

	"github.com/own-home/central/pkg/registry/driver"
	"github.com/own-home/central/pkg/spec"
)

// Registry is used to store, retrieve and inspect thing definitions
type Registry interface {
	// All returns all things stored at the registry
	All(context.Context) ([]*spec.Thing, error)

	// Get returns the thing by ID
	Get(context.Context, string) (*spec.Thing, error)

	// Create creates a new thing. The ID of the thing to create
	// must not yet exist
	Create(context.Context, *spec.Thing) error

	// Update updates an existing thing. The thing must already
	// exist. It's not allowed to change the ID of a thing
	Update(context.Context, *spec.Thing) error

	// Delete a thing by ID
	Delete(context.Context, string) error

	// ItemValues allows to store and retrieve item values
	ItemValues(context.Context, string, string) (driver.ValueStore, error)

	GetItemValue(context.Context, string, string) (interface{}, error)

	// RegisterCreatedNotifier registeres a notifier function that will be
	// called whenever a new thing is created
	RegisterCreatedNotifier(func(*spec.Thing))

	// RegisterUpdatedNotifier registers a notifier function that will be
	// called whenever an existing thing is updated
	RegisterUpdatedNotifier(func(*spec.Thing))

	// RegisterDeletedNotifier registers a notifier function that will be
	// called whenever an existing thing has been deleted
	RegisterDeletedNotifier(func(*spec.Thing))
}

// Open opens the registry using the provided driver name and options.
// For the exact format of the options please refer to the driver you
// want to use
func Open(driverName, driverOptions string) (Registry, error) {
	drv, err := driver.OpenDriver(driverName, driverOptions)
	if err != nil {
		return nil, err
	}

	return &registry{
		drv: drv,
	}, nil
}

type registry struct {
	drv driver.Driver

	notifiers        sync.RWMutex
	createdNotifiers []func(*spec.Thing)
	updatedNotifiers []func(*spec.Thing)
	deletedNotifiers []func(*spec.Thing)
}

// All returns all things stored in the registry
func (r *registry) All(ctx context.Context) ([]*spec.Thing, error) {
	ids, err := r.drv.IDs((ctx))
	if err != nil {
		return nil, err
	}

	things := make([]*spec.Thing, len(ids))

	for i, id := range ids {
		t, err := r.drv.Get(ctx, id)
		if err != nil {
			// seem like the thing has been deleted in between
			if err == driver.ErrUnknownThing {
				continue
			}

			return nil, err
		}

		things[i] = t
	}

	return things, nil
}

// Get returns a single thing defintion
func (r *registry) Get(ctx context.Context, id string) (*spec.Thing, error) {
	return r.drv.Get(ctx, id)
}

// Create creates a new thing definition and stores it in the registry
func (r *registry) Create(ctx context.Context, thing *spec.Thing) error {
	err := r.drv.Set(ctx, thing, &driver.SetOptions{
		CreateOnly: true,
	})

	if err == nil {
		go r.notifyCreated(thing)
	}

	return err
}

// Update an existing thing definition inside the registry
func (r *registry) Update(ctx context.Context, thing *spec.Thing) error {
	err := r.drv.Set(ctx, thing, &driver.SetOptions{
		UpdateOnly: true,
	})

	if err == nil {
		go r.notifyUpdated(thing)
	}

	return err
}

// Delete a thing definition from the registry
func (r *registry) Delete(ctx context.Context, id string) error {
	t, err := r.drv.Delete(ctx, id, &driver.DeleteOptions{
		MustExist: true,
	})

	if err == nil {
		go r.notifyDeleted(t)
	}

	return err
}

func (r *registry) ItemValues(ctx context.Context, thingID, itemID string) (driver.ValueStore, error) {
	return r.drv.ItemValues(ctx, thingID, itemID)
}

func (r *registry) GetItemValue(ctx context.Context, thingID, itemID string) (interface{}, error) {
	store, err := r.ItemValues(ctx, thingID, itemID)
	if err != nil {
		return nil, err
	}

	return store.Current(ctx)
}

// RegisterCreatedNotifier registers a new notifier function to be called
// when new things get registered
func (r *registry) RegisterCreatedNotifier(fn func(*spec.Thing)) {
	r.notifiers.Lock()
	defer r.notifiers.Unlock()

	r.createdNotifiers = append(r.createdNotifiers, fn)
}

// RegisterUpdatedNotifier registeres a new notifier function to be called
// when a thing definition gets updated
func (r *registry) RegisterUpdatedNotifier(fn func(*spec.Thing)) {
	r.notifiers.Lock()
	defer r.notifiers.Unlock()

	r.updatedNotifiers = append(r.updatedNotifiers, fn)
}

// RegisterDeletedNotifier registers a new notifier function to be called
// when a thing definition gets removed
func (r *registry) RegisterDeletedNotifier(fn func(*spec.Thing)) {
	r.notifiers.Lock()
	defer r.notifiers.Unlock()

	r.deletedNotifiers = append(r.deletedNotifiers, fn)
}

func (r *registry) notifyCreated(t *spec.Thing) {
	r.notifiers.RLock()
	defer r.notifiers.RUnlock()

	for _, fn := range r.createdNotifiers {
		go fn(t)
	}
}

func (r *registry) notifyUpdated(t *spec.Thing) {
	r.notifiers.RLock()
	defer r.notifiers.RUnlock()

	for _, fn := range r.updatedNotifiers {
		go fn(t)
	}
}

func (r *registry) notifyDeleted(t *spec.Thing) {
	r.notifiers.RLock()
	defer r.notifiers.RUnlock()

	for _, fn := range r.deletedNotifiers {
		go fn(t)
	}
}
