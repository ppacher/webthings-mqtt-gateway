package driver

import (
	"context"
	"sync"

	"github.com/ppacher/webthings-mqtt-gateway/pkg/spec"
)

// SetOptions alter the behavior of the set operation
type SetOptions struct {
	UpdateOnly bool
	CreateOnly bool
}

// DeleteOptions alter the behavior of the delete operation
type DeleteOptions struct {
	MustExist bool
}

// Driver is a storage driver for thing and item definitions
type Driver interface {
	// Get should return the thing identified by name or an appropriate
	// error
	Get(context.Context, string) (*spec.Thing, error)

	// Set updates a complete thing definition
	Set(context.Context, *spec.Thing, *SetOptions) error

	// Delete deletes a thing from the storage
	Delete(context.Context, string, *DeleteOptions) (*spec.Thing, error)

	// Has should return true if the provided ID is known
	Has(context.Context, string) (bool, error)

	// IDs returns a slice of registered thing IDs
	IDs(context.Context) ([]string, error)

	// ItemValues returns a ValueStore for the given thing item
	ItemValues(context.Context, string, string) (ValueStore, error)
}

// Factory is used to create a new driver object based on the given configuration
// string
type Factory func(options string) (Driver, error)

var factoriesLock sync.RWMutex
var factories = map[string]Factory{}

// Register registers a new driver factory for the given name. Any error
// is returned
func Register(name string, d Factory) error {
	factoriesLock.Lock()
	defer factoriesLock.Unlock()

	if _, ok := factories[name]; ok {
		return ErrDriverAlreadyRegistered
	}

	factories[name] = d

	return nil
}

// MustRegister registers the Factory d under the name n. It panics
// if there is an error. See Register for more information
func MustRegister(n string, d Factory) {
	if err := Register(n, d); err != nil {
		panic(err)
	}
}

// OpenDriver opens the driver with the given name
func OpenDriver(n string, options string) (Driver, error) {
	factoriesLock.RLock()
	defer factoriesLock.RUnlock()

	factory, ok := factories[n]
	if !ok {
		return nil, ErrUnknownDriver
	}

	return factory(options)
}
