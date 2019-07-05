package payload

import (
	"fmt"
	"sync"
)

// HandlerType identifies a handler type by name
type HandlerType string

// UnknownHandler is returned when the payload handler is unknown
const UnknownHandler = HandlerType("")

// Handler is capable of parsing and extracting data from a given payload
type Handler interface {
	// Parse should parse the payload and return the extracted value
	// or an error
	Parse(payload []byte, cfg HandlerSpec) (interface{}, error)
}

var handlers map[HandlerType]Handler
var handlersLock sync.RWMutex

// RegisterType registers a new handler type. Each handler type must have
// a unique name
func RegisterType(name HandlerType, h Handler) error {
	handlersLock.Lock()
	defer handlersLock.Unlock()

	if handlers == nil {
		handlers = make(map[HandlerType]Handler)
	}

	if _, ok := handlers[name]; ok {
		return ErrAlreadyRegistered
	}

	handlers[name] = h

	return nil
}

func isValid(v string) bool {
	handlersLock.RLock()
	defer handlersLock.RUnlock()

	if handlers == nil {
		return false
	}

	_, ok := handlers[HandlerType(v)]
	return ok
}

// MustRegisterType registers a new handler type. It panics if the handler
// type is already registered
func MustRegisterType(name HandlerType, h Handler) {
	if err := RegisterType(name, h); err != nil {
		panic(err)
	}
}

// HandlerSpec defines the handler used to parse a given payload
type HandlerSpec map[string]interface{}

// Handler returns the handler specified by this spec
func (h HandlerSpec) Handler() (Handler, error) {
	t, ok := h["type"]
	if !ok {
		return nil, ErrNoType
	}

	v, ok := t.(string)
	if !ok {
		return nil, ErrInvalidType
	}

	handlersLock.RLock()
	defer handlersLock.RUnlock()

	if handlers == nil {
		return nil, ErrInvalidType
	}

	handler, ok := handlers[HandlerType(v)]
	if !ok {
		return nil, ErrInvalidType
	}

	return handler, nil
}

// Type returns the type of the handler
func (h HandlerSpec) Type() (HandlerType, error) {
	t, ok := h["type"]
	if !ok {
		return UnknownHandler, ErrNoType
	}

	v, ok := t.(string)
	if !ok || !isValid(v) {
		return UnknownHandler, ErrInvalidType
	}

	return HandlerType(v), nil
}

// Parse tries to parse the given payload
func (h HandlerSpec) Parse(payload []byte) (res interface{}, err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", x)
			}
		}
	}()

	handler, err := h.Handler()
	if err != nil {
		return nil, err
	}

	res, err = handler.Parse(payload, h)

	return res, err
}

// GetInt returns an integer from the HandlerSpec with the specified key
// The return values represent the number, if the key was set and if the
// key was actually a number.
//
// Example:
//
//		value, keyWasPresent, keyWasNumber := h.GetInt("port")
//
func (h HandlerSpec) GetInt(key string) (int, bool, bool) {
	v, ok := h[key]
	if !ok {
		return 0, false, false
	}

	n, ok := v.(float64)
	if ok {
		return int(n), true, true
	}

	i, ok := v.(int)
	if ok {
		return i, true, true
	}

	return 0, true, false
}
