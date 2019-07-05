package driver

import (
	"net/http"

	"github.com/own-home/central/pkg/errors"
)

var (
	// ErrThingExists is returns if the given thing ID already exists
	ErrThingExists = errors.NewWithStatus(http.StatusConflict, "thing ID already exists")

	// ErrUnknownThing indicates that the request thing does not exist
	ErrUnknownThing = errors.NewWithStatus(http.StatusNotFound, "unknown thing")

	// ErrDriverAlreadyRegistered indicates that a driver with the same name has already
	// been registered
	ErrDriverAlreadyRegistered = errors.NewWithStatus(http.StatusConflict, "driver name already registered")

	// ErrUnknownDriver indicates that the requested driver is not registered
	ErrUnknownDriver = errors.NewWithStatus(http.StatusNotFound, "unknown driver. Did you forget to import it?")

	// ErrInvalidOptions is returned if a driver encounters invalid operation options
	ErrInvalidOptions = errors.NewWithStatus(http.StatusBadRequest, "invalid options passed to the driver")
)
